package schedule

import (
	"errors"
	"fmt"

	"github.com/creativeprojects/clog"
)

//
// Job: common code for all scheduling systems
//

var ErrJobCanBeRemovedOnly = errors.New("this job is marked for removal only and cannot be created or modified")

// Job scheduler
type Job struct {
	config  *Config
	handler Handler
}

func NewJob(handler Handler, config *Config) *Job {
	if handler == nil {
		panic("NewJob: handler cannot be nil")
	}
	if config == nil {
		panic("NewJob: config cannot be nil")
	}
	return &Job{
		config:  config,
		handler: handler,
	}
}

// Accessible checks if the current user is permitted to access the job
func (j *Job) Accessible() bool {
	permission, _ := j.handler.DetectSchedulePermission(PermissionFromConfig(j.config.Permission))
	return j.handler.CheckPermission(permission)
}

// Create a new job
func (j *Job) Create() error {
	if j.RemoveOnly() {
		return ErrJobCanBeRemovedOnly
	}

	permission := j.getSchedulePermission(PermissionFromConfig(j.config.Permission))
	if ok := j.handler.CheckPermission(permission); !ok {
		return permissionError("create")
	}

	if err := j.handler.DisplaySchedules(j.config.ProfileName, j.config.CommandName, j.config.Schedules); err != nil {
		return err
	}

	schedules, err := j.handler.ParseSchedules(j.config.Schedules)
	if err != nil {
		return err
	}

	if err = j.handler.CreateJob(j.config, schedules, permission); err != nil {
		return err
	}

	return nil
}

// Remove a job
func (j *Job) Remove() error {
	permission := PermissionFromConfig(j.config.Permission)
	if j.RemoveOnly() {
		permission, _ = j.handler.DetectSchedulePermission(permission) // silent call for possibly non-existent job
	} else {
		permission = j.getSchedulePermission(permission)
	}
	if ok := j.handler.CheckPermission(permission); !ok {
		return permissionError("remove")
	}

	return j.handler.RemoveJob(j.config, permission)
}

// RemoveOnly returns true if this job can be removed only
func (j *Job) RemoveOnly() bool {
	return j.config.removeOnly
}

// Status of a job
func (j *Job) Status() error {
	if j.RemoveOnly() {
		return ErrJobCanBeRemovedOnly
	}

	if err := j.handler.DisplaySchedules(j.config.ProfileName, j.config.CommandName, j.config.Schedules); err != nil {
		return err
	}

	if err := j.handler.DisplayJobStatus(j.config); err != nil {
		return err
	}
	return nil
}

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// If the permission can only be guessed, this method will also display a warning
func (j *Job) getSchedulePermission(permission Permission) Permission {
	permission, safe := j.handler.DetectSchedulePermission(permission)
	if !safe {
		clog.Warningf("you have not specified the permission for your schedule (\"system\", \"user\" or \"user_logged_on\"): assuming %q", permission.String())
	}
	return permission
}

// permissionError display a permission denied message to the user.
// permissionError is not used in Windows.
func permissionError(action string) error {
	return fmt.Errorf("user is not allowed to %s a system job: please restart resticprofile as root (with sudo)", action)
}
