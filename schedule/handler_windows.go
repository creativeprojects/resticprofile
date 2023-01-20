//go:build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schtasks"
)

// HandlerWindows is using windows task manager
type HandlerWindows struct {
	config SchedulerConfig
}

// NewHandler creates a new handler for windows task manager
func NewHandler(config SchedulerConfig) *HandlerWindows {
	return &HandlerWindows{
		config: config,
	}
}

// Init a connection to the task scheduler
func (h *HandlerWindows) Init() error {
	return schtasks.Connect()
}

// Close the connection to the task scheduler
func (h *HandlerWindows) Close() {
	schtasks.Close()
}

// ParseSchedules into *calendar.Event
func (h *HandlerWindows) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

// DisplayParsedSchedules via term output
func (h *HandlerWindows) DisplayParsedSchedules(command string, events []*calendar.Event) {
	displayParsedSchedules(command, events)
}

// DisplaySchedules does nothing on windows
func (h *HandlerWindows) DisplaySchedules(command string, schedules []string) error {
	return nil
}

// DisplayStatus does nothing on windows task manager
func (h *HandlerWindows) DisplayStatus(profileName string) error {
	return nil
}

// CreateJob is creating the task scheduler job.
func (h *HandlerWindows) CreateJob(job *config.ScheduleConfig, schedules []*calendar.Event, permission string) error {
	// default permission will be system
	perm := schtasks.SystemAccount
	if permission == constants.SchedulePermissionUser {
		perm = schtasks.UserAccount
	} else if permission == constants.SchedulePermissionUserLoggedOn {
		perm = schtasks.UserLoggedOnAccount
	}
	err := schtasks.Create(job, schedules, perm)
	if err != nil {
		return err
	}
	return nil
}

// RemoveJob is deleting the task scheduler job
func (h *HandlerWindows) RemoveJob(job *config.ScheduleConfig, permission string) error {
	err := schtasks.Delete(job.Title, job.SubTitle)
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}

// DisplayStatus display some information about the task scheduler job
func (h *HandlerWindows) DisplayJobStatus(job *config.ScheduleConfig) error {
	err := schtasks.Status(job.Title, job.SubTitle)
	if err != nil {
		if errors.Is(err, schtasks.ErrorNotRegistered) {
			return ErrorServiceNotFound
		}
		return err
	}
	return nil
}
