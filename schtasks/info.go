package schtasks

import (
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"os/exec"
	"strings"
)

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

func getCSV(input io.Reader) ([][]string, error) {
	reader := csv.NewReader(input)
	return reader.ReadAll()
}

func schTasksError(message string) error {
	message = strings.TrimSpace(message)
	message = strings.TrimPrefix(message, "ERROR: ")
	if strings.Contains(message, "The system cannot find the file specified") ||
		strings.Contains(message, "does not exist in the system") {
		return ErrNotRegistered
	} else if strings.Contains(message, "The filename, directory name, or volume label syntax is incorrect") {
		return ErrInvalidTaskName
	} else if strings.Contains(message, "Access is denied") {
		return ErrAccessDenied
	}
	return errors.New(message)
}
