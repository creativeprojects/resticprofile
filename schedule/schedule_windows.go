//+build windows

package schedule

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/capnspacehook/taskmaster"
	"golang.org/x/sys/windows"
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
		// return retryElevated(err)
		return err
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

	return taskService.DeleteTask(getTaskPath(profileName))
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
		// return retryElevated(err)
		return err
	}
	if registeredTask == nil {
		return fmt.Errorf("task '%s' is not registered in the task scheduler", taskName)
	}
	fmt.Printf("%+v", registeredTask)
	return nil
}

func retryElevated(err error) error {
	if strings.Contains(err.Error(), "denied") {
		err := runElevated()
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// runElevated restart resticprofile in elevated mode.
// it is not in use right now as it opens a new console window.
// I need to find a way to attach the new process to the existing console window.
func runElevated() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(getConsoleHandle(), verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}
	return nil
}

func getConsoleHandle() windows.Handle {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	ret, _, _ := getConsoleWindow.Call()
	return windows.Handle(ret)
}
