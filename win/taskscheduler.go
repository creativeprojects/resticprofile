//+build windows

package win

import (
	"fmt"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/resticprofile/config"
)

const (
	tasksPath = `\resticprofile backup\`
)

type TaskScheduler struct {
	profile *config.Profile
}

func NewTaskScheduler(profile *config.Profile) *TaskScheduler {
	return &TaskScheduler{
		profile: profile,
	}
}

func (s *TaskScheduler) Create(binary, args, workingDir, description string) error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	task := taskService.NewTaskDefinition()
	task.AddExecAction(binary, args, workingDir, "")
	task.Principal.LogonType = taskmaster.TASK_LOGON_SERVICE_ACCOUNT
	task.Principal.RunLevel = taskmaster.TASK_RUNLEVEL_HIGHEST
	task.Principal.UserID = "SYSTEM"
	task.RegistrationInfo.Author = "resticprofile"
	task.RegistrationInfo.Description = description
	_, _, err = taskService.CreateTask(getTaskPath(s.profile.Name), task, true)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskScheduler) Update() error {
	return nil
}

func (s *TaskScheduler) Delete() error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	err = taskService.DeleteTask(getTaskPath(s.profile.Name))
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskScheduler) Status() error {
	taskService, err := s.connect()
	if err != nil {
		return err
	}
	defer taskService.Disconnect()

	taskName := getTaskPath(s.profile.Name)
	registeredTask, err := taskService.GetRegisteredTask(taskName)
	if err != nil {
		return err
	}
	if registeredTask == nil {
		return fmt.Errorf("task '%s' is not registered in the task scheduler", taskName)
	}
	fmt.Printf("%+v", registeredTask)
	return nil
}

func (s *TaskScheduler) connect() (*taskmaster.TaskService, error) {
	return taskmaster.Connect("", "", "", "")
}

func getTaskPath(profileName string) string {
	return tasksPath + profileName
}
