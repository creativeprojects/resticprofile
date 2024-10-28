//go:build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schtasks"
)

// HandlerWindows is using windows task manager
type HandlerWindows struct {
	config SchedulerConfig
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

// DisplaySchedules via term output
func (h *HandlerWindows) DisplaySchedules(profile, command string, schedules []string) error {
	events, err := parseSchedules(schedules)
	if err != nil {
		return err
	}
	displayParsedSchedules(profile, command, events)
	return nil
}

// DisplayStatus does nothing on windows task manager
func (h *HandlerWindows) DisplayStatus(profileName string) error {
	return nil
}

// CreateJob is creating the task scheduler job.
func (h *HandlerWindows) CreateJob(job *Config, schedules []*calendar.Event, permission string) error {
	// default permission will be system
	perm := schtasks.SystemAccount
	if permission == constants.SchedulePermissionUser {
		perm = schtasks.UserAccount
	} else if permission == constants.SchedulePermissionUserLoggedOn || permission == constants.SchedulePermissionUserLoggedIn {
		perm = schtasks.UserLoggedOnAccount
	}
	jobConfig := &schtasks.Config{
		ProfileName:      job.ProfileName,
		CommandName:      job.CommandName,
		Command:          job.Command,
		Arguments:        job.Arguments.String(),
		WorkingDirectory: job.WorkingDirectory,
		JobDescription:   job.JobDescription,
	}
	err := schtasks.Create(jobConfig, schedules, perm)
	if err != nil {
		return err
	}
	return nil
}

// RemoveJob is deleting the task scheduler job
func (h *HandlerWindows) RemoveJob(job *Config, permission string) error {
	err := schtasks.Delete(job.ProfileName, job.CommandName)
	if err != nil {
		if errors.Is(err, schtasks.ErrNotRegistered) {
			return ErrServiceNotFound
		}
		return err
	}
	return nil
}

// DisplayStatus display some information about the task scheduler job
func (h *HandlerWindows) DisplayJobStatus(job *Config) error {
	err := schtasks.Status(job.ProfileName, job.CommandName)
	if err != nil {
		if errors.Is(err, schtasks.ErrNotRegistered) {
			return ErrServiceNotFound
		}
		return err
	}
	return nil
}

// init registers HandlerWindows
func init() {
	AddHandlerProvider(func(config SchedulerConfig, _ bool) (hr Handler) {
		if config.Type() == constants.SchedulerWindows ||
			config.Type() == constants.SchedulerOSDefault {
			hr = &HandlerWindows{
				config: config,
			}
		}
		return
	})
}
