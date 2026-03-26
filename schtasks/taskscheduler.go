//go:build windows

// Package schtasks
//
// Schedule types on Windows:
// ==========================
// 1. one time:
//   - at a specific date
//
// 2. daily:
//   - 1 start date
//   - recurring every n days
//
// 3. weekly:
//   - 1 start date
//   - recurring every n weeks
//   - on specific weekdays
//
// 4. monthly:
//   - 1 start date
//   - on specific months
//   - on specific days (1 to 31)
package schtasks

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
)

// Create or update a task (if the name already exists in the Task Scheduler)
func Create(config *Config, schedules []*calendar.Event, permission Permission) error {
	var (
		username, password string
		err                error
	)
	list, err := getRegisteredTasks()
	if err != nil {
		return fmt.Errorf("error loading list of registered tasks: %w", err)
	}

	taskPath := getTaskPath(config.ProfileName, config.CommandName)
	if slices.Contains(list, taskPath) {
		clog.Debugf("task %q already exists: deleting before creating", taskPath)
		_, err = deleteTask(taskPath)
		if err != nil {
			return fmt.Errorf("cannot delete existing task to replace it: %w", err)
		}
	}

	task := createTaskDefinition(config, schedules)
	task.RegistrationInfo.URI = taskPath

	switch config.RunLevel {
	case "lowest":
		task.Principals.Principal.RunLevel = RunLevelLeastPrivilege

	case "highest":
		task.Principals.Principal.RunLevel = RunLevelHighest

	default:
		if permission == SystemAccount {
			task.Principals.Principal.RunLevel = RunLevelHighest
		} else {
			task.Principals.Principal.RunLevel = RunLevelDefault
		}
	}

	switch permission {
	case SystemAccount:
		task.Principals.Principal.LogonType = LogonTypeServiceForUser
		task.Principals.Principal.UserId = serviceAccount
		task.RegistrationInfo.SecurityDescriptor = securityDescriptor // allow authenticated users to read the task status

	case UserLoggedOnAccount:
		task.Principals.Principal.LogonType = LogonTypeInteractiveToken

	default:
		task.Principals.Principal.LogonType = LogonTypePassword
		username, password, err = userCredentials()
		if err != nil {
			return fmt.Errorf("cannot get user name or password: %w", err)
		}
	}

	file, err := os.CreateTemp("", "*.xml")
	if err != nil {
		return fmt.Errorf("cannot create XML task file: %w", err)
	}
	filename := file.Name()
	defer func() {
		file.Close()
		os.Remove(filename)
	}()

	err = createTaskFile(task, file)
	if err != nil {
		return fmt.Errorf("cannot write XML task file: %w", err)
	}
	file.Close()

	_, err = createTask(taskPath, filename, username, password)
	return err
}

// Delete a task
func Delete(title, subtitle string) error {
	taskName := getTaskPath(title, subtitle)
	_, err := deleteTask(taskName)
	return err
}

// Status returns the status of a task
func Status(title, subtitle string) error {
	taskName := getTaskPath(title, subtitle)
	info, err := getTaskInfo(taskName)
	if err != nil {
		return err
	}
	if len(info) < 2 {
		return ErrNotRegistered
	}
	writer := tabwriter.NewWriter(term.GetOutput(), 2, 2, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(writer, "Task:\t %s\n", getFirstField(info, "TaskName"))
	fmt.Fprintf(writer, "User:\t %s\n", getFirstField(info, "Run As User"))
	fmt.Fprintf(writer, "Logon Mode:\t %s\n", getFirstField(info, "Logon Mode"))
	fmt.Fprintf(writer, "Working Dir:\t %v\n", getFirstField(info, "Start In"))
	fmt.Fprintf(writer, "Command:\t %v\n", getFirstField(info, "Task To Run"))
	fmt.Fprintf(writer, "Status:\t %s\n", getFirstField(info, "Status"))
	fmt.Fprintf(writer, "Last Run Time:\t %v\n", getFirstField(info, "Last Run Time"))
	fmt.Fprintf(writer, "Last Result:\t %s\n", getFirstField(info, "Last Result"))
	fmt.Fprintf(writer, "Next Run Time:\t %v\n", getFirstField(info, "Next Run Time"))
	writer.Flush()
	return nil
}

func Registered() ([]Config, error) {
	list, err := getRegisteredTasks()
	if err != nil {
		return nil, fmt.Errorf("error loading list of registered tasks: %w", err)
	}

	configs := make([]Config, 0, len(list))
	for _, taskPath := range list {
		clog.Debugf("loading task %q", taskPath)
		info, err := getTaskInfo(taskPath)
		if err != nil {
			clog.Errorf("loading task %q: %s", taskPath, err)
			continue
		}
		taskName := strings.TrimPrefix(taskPath, tasksPathPrefix)
		parts := strings.Split(taskName, " ")
		if len(parts) < 2 {
			clog.Warningf("cannot parse task path: %s", taskPath)
			continue
		}
		profileName := strings.Join(parts[:len(parts)-1], " ")
		commandName := parts[len(parts)-1]
		commandLine := getFirstField(info, "Task To Run")
		command, args, _ := strings.Cut(commandLine, " ")
		config := Config{
			ProfileName:      profileName,
			CommandName:      commandName,
			JobDescription:   getFirstField(info, "Comment"),
			WorkingDirectory: getFirstField(info, "Start In"),
			Command:          command,
			Arguments:        args,
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func getTaskPath(profileName, commandName string) string {
	return fmt.Sprintf("%s%s %s", tasksPathPrefix, profileName, commandName)
}

func createTaskDefinition(config *Config, schedules []*calendar.Event) Task {
	task := NewTask()
	task.RegistrationInfo.Description = config.JobDescription
	task.Settings.StartWhenAvailable = config.StartWhenAvailable
	task.AddExecAction(ExecAction{
		Command:          config.Command,
		Arguments:        config.Arguments,
		WorkingDirectory: config.WorkingDirectory,
	})
	task.AddSchedules(schedules)
	return task
}
