//+build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/win"
)

// Init a connection to the task scheduler
func Init() error {
	return win.ConnectScheduler()
}

// Close the connection to the task scheduler
func Close() {
	win.CloseScheduler()
}

// createJob is creating the task scheduler job.
func (j *Job) createJob(schedules []*calendar.Event) error {
	// default permission will be system
	permission := win.SystemAccount
	if j.config.Permission() == constants.SchedulePermissionUser {
		permission = win.UserAccount
	}
	taskScheduler := win.NewTaskScheduler(j.config)
	err := taskScheduler.Create(schedules, permission)
	if err != nil {
		return err
	}
	return nil
}

// removeJob is deleting the task scheduler job
func (j *Job) removeJob() error {
	taskScheduler := win.NewTaskScheduler(j.config)
	err := taskScheduler.Delete()
	if err != nil {
		if errors.Is(err, win.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// displayStatus display some information about the task scheduler job
func (j *Job) displayStatus(command string) error {
	taskScheduler := win.NewTaskScheduler(j.config)
	err := taskScheduler.Status()
	if err != nil {
		if errors.Is(err, win.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}
