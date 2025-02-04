package schedule

import (
	"errors"
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
	permission, _ := detectSchedulePermission(j.config.Permission)
	return checkPermission(permission)
}

// Create a new job
func (j *Job) Create() error {
	if j.RemoveOnly() {
		return ErrJobCanBeRemovedOnly
	}

	permission := getSchedulePermission(j.config.Permission)
	ok := checkPermission(permission)
	if !ok {
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
	var permission string
	if j.RemoveOnly() {
		permission, _ = detectSchedulePermission(j.config.Permission) // silent call for possibly non-existent job
	} else {
		permission = getSchedulePermission(j.config.Permission)
	}
	ok := checkPermission(permission)
	if !ok {
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
