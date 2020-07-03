//+build windows

package schedule

import (
	"errors"
	"fmt"
	"os"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/resticprofile/calendar"
)

const (
	tasksPath = `\resticprofile backup\`
)

// createJob is creating the task manager job.
func (j *Job) createJob() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	taskService, err := taskmaster.Connect("", "", "", "")
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	// office, err := taskService.GetRegisteredTask("\\Microsoft\\Office\\Office Automatic Updates 2.0")
	// if err != nil {
	// 	return err
	// }
	// fmt.Printf("%+v", office)

	binary := absolutePathToBinary(wd, os.Args[0])

	task := taskService.NewTaskDefinition()
	task.AddExecAction(binary, fmt.Sprintf("--no-ansi --config %s --name %s backup", j.configFile, j.profile.Name), wd, "")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Description = fmt.Sprintf("restic backup using profile '%s' from '%s'", j.profile.Name, j.configFile)
	_, _, err = taskService.CreateTask(tasksPath+j.profile.Name, task, true)
	if err != nil {
		return err
	}
	return nil
}

func RemoveJob(profileName string) error {
	taskService, err := taskmaster.Connect("", "", "", "")
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	return taskService.DeleteTask(tasksPath + profileName)
}

func (j *Job) displayStatus() error {
	return nil
}

func loadSchedules(schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing schedule %d/%d\n========================\n", index+1, len(schedules))
	}
	return events, nil
}
