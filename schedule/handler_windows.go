//go:build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schtasks"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/user"
)

// HandlerWindows is using windows task manager
type HandlerWindows struct {
	config SchedulerConfig
}

// Init only checks the schtask.exe tool is available
func (h *HandlerWindows) Init() error {
	return lookupBinary("schtasks", "schtasks.exe")
}

// Close does nothing with this implementation
func (h *HandlerWindows) Close() {
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
func (h *HandlerWindows) CreateJob(job *Config, schedules []*calendar.Event, permission Permission) error {
	// default permission will be system
	perm := schtasks.SystemAccount
	switch permission {
	case PermissionUserBackground:
		perm = schtasks.UserAccount
	case PermissionUserLoggedOn:
		perm = schtasks.UserLoggedOnAccount
	}

	var command string
	var arguments CommandArguments

	if job.HideWindow {
		if permission != PermissionUserLoggedOn {
			clog.Warning("hiding window makes sense only with \"user_logged_on\" permission")
		}

		command = "conhost.exe"
		arguments = NewCommandArguments(append(
			[]string{"--headless", job.Command},
			job.Arguments.RawArgs()...,
		))
	} else {
		command = job.Command
		arguments = job.Arguments
	}

	jobConfig := &schtasks.Config{
		ProfileName:        job.ProfileName,
		CommandName:        job.CommandName,
		Command:            command,
		Arguments:          arguments.String(),
		WorkingDirectory:   job.WorkingDirectory,
		JobDescription:     job.JobDescription,
		RunLevel:           job.RunLevel,
		StartWhenAvailable: job.StartWhenAvailable,
	}
	err := schtasks.Create(jobConfig, schedules, perm)
	if err != nil {
		return err
	}
	return nil
}

// RemoveJob is deleting the task scheduler job
func (h *HandlerWindows) RemoveJob(job *Config, _ Permission) error {
	err := schtasks.Delete(job.ProfileName, job.CommandName)
	if err != nil {
		if errors.Is(err, schtasks.ErrNotRegistered) {
			return ErrScheduledJobNotFound
		}
		return err
	}
	return nil
}

// DisplayJobStatus display some information about the task scheduler job
func (h *HandlerWindows) DisplayJobStatus(job *Config) error {
	err := schtasks.Status(job.ProfileName, job.CommandName)
	if err != nil {
		if errors.Is(err, schtasks.ErrNotRegistered) {
			return ErrScheduledJobNotFound
		}
		return err
	}
	return nil
}

func (h *HandlerWindows) Scheduled(profileName string) ([]Config, error) {
	tasks, err := schtasks.Registered()
	if err != nil {
		return nil, err
	}
	configs := make([]Config, 0, len(tasks))
	for _, task := range tasks {
		if profileName == "" || task.ProfileName == profileName {
			args := NewCommandArguments(shell.SplitArguments(task.Arguments))
			configs = append(configs, Config{
				ConfigFile:       args.ConfigFile(),
				ProfileName:      task.ProfileName,
				CommandName:      task.CommandName,
				Command:          task.Command,
				Arguments:        args,
				WorkingDirectory: task.WorkingDirectory,
				JobDescription:   task.JobDescription,
			})
		}
	}
	return configs, nil
}

// DetectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func (h *HandlerWindows) DetectSchedulePermission(permission Permission) (Permission, bool) {
	switch permission {
	case PermissionAuto:
		return PermissionSystem, true

	default:
		return permission, true
	}
}

// CheckPermission returns true if the user is allowed to access the job.
// This is always true on Windows.
func (h *HandlerWindows) CheckPermission(_ user.User, _ Permission) bool {
	return true
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
