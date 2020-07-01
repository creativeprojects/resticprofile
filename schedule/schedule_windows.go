//+build windows

package schedule

import (
	"fmt"
	"os"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/resticprofile/config"
)

const (
	tasksPath = `\resticprofile backup\`
)

func CreateJob(configFile string, profile *config.Profile) error {
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
	task.AddExecAction(binary, fmt.Sprintf("--no-ansi --config %s --name %s backup", configFile, profile.Name), wd, "")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Description = fmt.Sprintf("restic backup using profile '%s' from '%s'", profile.Name, configFile)
	_, _, err = taskService.CreateTask(tasksPath+profile.Name, task, true)
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
