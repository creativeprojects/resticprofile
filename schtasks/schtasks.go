package schtasks

import (
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
)

func createTaskDefinition(config *Config, schedules []*calendar.Event) Task {
	task := NewTask()
	task.RegistrationInfo.Description = config.JobDescription
	task.AddExecAction(ExecAction{
		Command:          config.Command,
		Arguments:        config.Arguments,
		WorkingDirectory: config.WorkingDirectory,
	})
	task.AddSchedules(schedules)
	return task
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

func createTask(config *Config, schedules []*calendar.Event) (string, *Task, error) {
	file, err := os.CreateTemp("", "*.xml")
	if err != nil {
		return "", nil, fmt.Errorf("cannot create XML task file: %w", err)
	}
	filename := file.Name()
	defer func() {
		file.Close()
		os.Remove(filename)
	}()

	taskPath := getTaskPath(config.ProfileName, config.CommandName)
	task := createTaskDefinition(config, schedules)
	task.RegistrationInfo.URI = taskPath

	err = createTaskFile(task, file)
	if err != nil {
		return "", nil, fmt.Errorf("cannot write XML task file: %w", err)
	}
	file.Close()

	buffer := &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/create", "/tn", taskPath, "/xml", filename)
	cmd.Stdout = buffer
	err = cmd.Run()
	return buffer.String(), &task, err
}

func importTaskDefinition(taskPath, filename string) error {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/create", "/tn", taskPath, "/xml", filename)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return schTasksError(stderr.String())
	}
	return nil
}

func exportTaskDefinition(taskName string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/query", "/xml", "/tn", taskName)
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.Bytes(), err
}

func listRegisteredTasks() ([]byte, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command(binaryPath, "/query", "/nh", "/fo", "csv")
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.Bytes(), err
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
