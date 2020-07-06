//+build windows

package schedule

import (
	"fmt"
	"os"

	"github.com/capnspacehook/taskmaster"
)

const (
	tasksPath = `\resticprofile backup\`
)

// createJob is creating the task scheduler job.
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

	binary := absolutePathToBinary(wd, os.Args[0])

	task := taskService.NewTaskDefinition()
	task.AddExecAction(binary, fmt.Sprintf("--no-ansi --config %s --name %s backup", j.configFile, j.profile.Name), wd, "")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = fmt.Sprintf("restic backup using profile '%s' from '%s'", j.profile.Name, j.configFile)
	_, _, err = taskService.CreateTask(getTaskPath(j.profile.Name), task, true)
	if err != nil {
		return retryElevated(err)
		// return err
	}
	return nil
}

func getTaskPath(profileName string) string {
	return tasksPath + profileName
}

func RemoveJob(profileName string) error {
	taskService, err := taskmaster.Connect("", "", "", "")
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	err = taskService.DeleteTask(getTaskPath(profileName))
	if err != nil {
		return retryElevated(err)
	}
	return nil
}

// checkSystem does nothing on windows as the task scheduler is always available
func checkSystem() error {
	return nil
}

func (j *Job) displayStatus() error {
	taskService, err := taskmaster.Connect("", "", "", "")
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	taskName := getTaskPath(j.profile.Name)
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return retryElevated(err)
		// return err
	}
	if registeredTask == nil {
		return fmt.Errorf("task '%s' is not registered in the task scheduler", taskName)
	}
	fmt.Printf("%+v", registeredTask)
	return nil
}

func retryElevated(err error) error {
	if err == nil {
		return nil
	}
	// maybe can find a better way than searching for the word "denied"?
	// if strings.Contains(err.Error(), "denied") {
	// 	clog.Info("restarting resticprofile in elevated mode...")
	// 	err := w32.RunElevated()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }
	return err
}
