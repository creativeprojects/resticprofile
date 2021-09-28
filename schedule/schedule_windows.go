//go:build windows
// +build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schtasks"
)

// Schedule using windows task manager
type Schedule struct {
}

// NewScheduler creates a Schedule onject (of Scheduler interface)
// On windows, only the task manager is supported
func NewScheduler(scheduler SchedulerType, profileName string) Scheduler {
	return &Schedule{}
}

// Init a connection to the task scheduler
func (s *Schedule) Init() error {
	return schtasks.Connect()
}

// Close the connection to the task scheduler
func (s *Schedule) Close() {
	schtasks.Close()
}

// NewJob instantiates a Job object (of SchedulerJob interface) to schedule jobs
func (s *Schedule) NewJob(config Config) SchedulerJob {
	return &Job{
		config: config,
	}
}

// DisplayStatus does nothing on windows task manager
func (s *Schedule) DisplayStatus() {
}

// Verify interface
var _ Scheduler = &Schedule{}

// createJob is creating the task scheduler job.
func (j *Job) createJob(schedules []*calendar.Event) error {
	// default permission will be system
	permission := schtasks.SystemAccount
	if p, _ := j.detectSchedulePermission(); p == constants.SchedulePermissionUser {
		permission = schtasks.UserAccount
	}
	err := schtasks.Create(j.config, schedules, permission)
	if err != nil {
		return err
	}
	return nil
}

// removeJob is deleting the task scheduler job
func (j *Job) removeJob() error {
	err := schtasks.Delete(j.config.Title(), j.config.SubTitle())
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// displayStatus display some information about the task scheduler job
func (j *Job) displayStatus(command string) error {
	err := schtasks.Status(j.config.Title(), j.config.SubTitle())
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// unsafe specifies whether a guess may lead to a too broad or too narrow file access permission.
//
// This method is for Windows only
func (j *Job) detectSchedulePermission() (permission string, unsafe bool) {
	if j.config.Permission() == constants.SchedulePermissionUser {
		return constants.SchedulePermissionUser, false
	}
	return constants.SchedulePermissionSystem, false
}

// checkPermission returns true if the user is allowed to access the job.
//
// This method is for Windows only
func (j *Job) checkPermission(permission string) bool {
	return true
}
