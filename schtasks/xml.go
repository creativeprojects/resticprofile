package schtasks

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
)

func createTaskDefinition(config *Config, schedules []*calendar.Event) (Task, error) {
	task := NewTask()
	task.RegistrationInfo.Description = config.JobDescription
	task.AddExecAction(ExecAction{
		Command:          config.Command,
		Arguments:        config.Arguments,
		WorkingDirectory: config.WorkingDirectory,
	})
	task.AddSchedules(schedules)
	return task, nil
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
	task, _ := createTaskDefinition(config, schedules)
	task.RegistrationInfo.URI = taskPath

	err = createTaskFile(task, file)
	if err != nil {
		return "", nil, fmt.Errorf("cannot write XML task file: %w", err)
	}
	file.Close()

	buffer := &bytes.Buffer{}
	cmd := exec.Command("schtasks", "/create", "/tn", taskPath, "/xml", filename)
	cmd.Stdout = buffer
	err = cmd.Run()
	return buffer.String(), &task, err
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPath, profileName, commandName)
}

func readTaskDefinition(taskName string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command("schtasks", "/query", "/xml", "/tn", taskName)
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.Bytes(), err
}

func deleteTask(taskName string) (string, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command("schtasks", "/delete", "/f", "/tn", taskName)
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.String(), err
}
