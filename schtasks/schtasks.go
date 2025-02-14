//go:build windows

package schtasks

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"io"
	"os/exec"
	"slices"
	"strings"
)

func getRegisteredTasks() ([]string, error) {
	raw, err := listRegisteredTasks()
	if err != nil {
		return nil, err
	}
	all, err := getCSV(bytes.NewBuffer(raw))
	if err != nil {
		return nil, err
	}
	list := make([]string, 0)
	for _, taskLine := range all {
		if len(taskLine) > 0 && strings.HasPrefix(taskLine[0], tasksPathPrefix) {
			list = append(list, taskLine[0])
		}
	}
	slices.Sort(list)
	list = slices.Compact(list)
	return list, nil
}

func getTaskInfo(taskName string) ([][]string, error) {
	buffer := &bytes.Buffer{}
	err := readTaskInfo(taskName, buffer)
	if err != nil {
		return nil, err
	}
	output, err := getCSV(buffer)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func getCSV(input io.Reader) ([][]string, error) {
	reader := csv.NewReader(input)
	return reader.ReadAll()
}

func createTaskFile(task Task, w io.Writer) error {
	var err error
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	err = encoder.Encode(&task)
	if err != nil {
		return err
	}
	return err
}

// createTask calls schtasks.exe to create a task from the XML file.
// username and password are optional, but must be specified at the same time.
func createTask(taskName, filename, username, password string) (string, error) {
	taskName = strings.TrimSpace(taskName)
	if len(taskName) == 0 {
		return "", ErrEmptyTaskName
	}
	params := []string{"/create", "/tn", taskName, "/xml", filename}

	if len(password) > 0 && len(username) == 0 {
		return "", errors.New("username is required when specifying a password")
	}
	if len(password) > 0 {
		params = append(params, "/ru", username, "/rp", password)
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(binaryPath, params...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), schTasksError(stderr.String())
	}
	return stdout.String(), nil
}

func exportTaskDefinition(taskName string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/query", "/xml", "/tn", taskName)
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.Bytes(), err
}

func listRegisteredTasks() ([]byte, error) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/query", "/nh", "/fo", "csv")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return stdout.Bytes(), schTasksError(stderr.String())
	}
	return stdout.Bytes(), err
}

func deleteTask(taskName string) (string, error) {
	taskName = strings.TrimSpace(taskName)
	if len(taskName) == 0 {
		return "", ErrEmptyTaskName
	}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/delete", "/f", "/tn", taskName)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return "", schTasksError(stderr.String())
	}
	return stdout.String(), nil
}

// readTaskInfo returns the raw CSV output from querying the task name (via schtasks.exe)
func readTaskInfo(taskName string, output io.Writer) error {
	taskName = strings.TrimSpace(taskName)
	if len(taskName) == 0 {
		return ErrEmptyTaskName
	}
	stderr := &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/query", "/fo", "csv", "/v", "/tn", taskName)
	cmd.Stdout = output
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return schTasksError(stderr.String())
	}
	return nil
}

func schTasksError(message string) error {
	message = strings.TrimSpace(message)
	message = strings.TrimPrefix(message, "ERROR: ")
	if strings.Contains(message, "The system cannot find the") ||
		strings.Contains(message, "does not exist in the system") {
		return ErrNotRegistered
	} else if strings.Contains(message, "The filename, directory name, or volume label syntax is incorrect") {
		return ErrInvalidTaskName
	} else if strings.Contains(message, "Access is denied") {
		return ErrAccessDenied
	} else if strings.Contains(message, "already exists") {
		return ErrAlreadyExist
	}
	return errors.New(message)
}
