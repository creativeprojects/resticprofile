//+build windows

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

// NewSchedule creates a Schedule onject (of Scheduler interface)
// On windows, only the task manager is supported
func NewSchedule(scheduler string) Scheduler {
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

// Verify interface
var _ Scheduler = &Schedule{}

// createJob is creating the task scheduler job.
func (j *Job) createJob(schedules []*calendar.Event) error {
	// default permission will be system
	permission := schtasks.SystemAccount
	if j.config.Permission() == constants.SchedulePermissionUser {
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
