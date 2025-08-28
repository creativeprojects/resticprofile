//go:build !windows

package schedule

import (
	"os"
	"slices"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/crond"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/user"
	"github.com/spf13/afero"
)

var crontabBinary = "crontab"

// HandlerCrond is a handler for crond scheduling
type HandlerCrond struct {
	config SchedulerCrond
	fs     afero.Fs
}

// NewHandlerCrond creates a new handler for crond scheduling
func NewHandlerCrond(config SchedulerConfig) *HandlerCrond {
	cfg, ok := config.(SchedulerCrond)
	if !ok {
		cfg = SchedulerCrond{}
	}
	if cfg.CrontabBinary == "" {
		cfg.CrontabBinary = platform.Executable(crontabBinary)
	}
	return &HandlerCrond{
		config: cfg,
		fs:     afero.NewOsFs(),
	}
}

// Init verifies crond is available on this system
func (h *HandlerCrond) Init() error {
	if len(h.config.CrontabFile) > 0 {
		clog.Debugf("using %q file as cron scheduler", h.config.CrontabFile)
		return nil
	}
	clog.Debug("using standard cron scheduler")
	return lookupBinary("crond", h.config.CrontabBinary)
}

// Close does nothing with crond
func (h *HandlerCrond) Close() {
	// nothing to do
}

func (h *HandlerCrond) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

func (h *HandlerCrond) DisplaySchedules(profile, command string, schedules []string) error {
	events, err := parseSchedules(schedules)
	if err != nil {
		return err
	}
	displayParsedSchedules(profile, command, events)
	return nil
}

// DisplayStatus does nothing with crond
func (h *HandlerCrond) DisplayStatus(profileName string) error {
	return nil
}

// CreateJob is creating the crontab
func (h *HandlerCrond) CreateJob(job *Config, schedules []*calendar.Event, permission Permission) error {
	entries := make([]crond.Entry, len(schedules))
	for i, event := range schedules {
		entries[i] = crond.NewEntry(
			event,
			job.ConfigFile,
			job.ProfileName,
			job.CommandName,
			job.Command+" "+job.Arguments.String(),
			job.WorkingDirectory,
		)
		if h.config.Username != "" {
			entries[i] = entries[i].WithUser(h.config.Username)
		}
	}
	crontab := crond.NewCrontab(entries).
		SetFile(h.config.CrontabFile).
		SetBinary(h.config.CrontabBinary).
		SetFs(h.fs)
	err := crontab.Rewrite()
	if err != nil {
		return err
	}
	return nil
}

func (h *HandlerCrond) RemoveJob(job *Config, permission Permission) error {
	entries := []crond.Entry{
		crond.NewEntry(
			calendar.NewEvent(),
			job.ConfigFile,
			job.ProfileName,
			job.CommandName,
			job.Command+" "+job.Arguments.String(),
			job.WorkingDirectory,
		),
	}
	crontab := crond.NewCrontab(entries).
		SetFile(h.config.CrontabFile).
		SetBinary(h.config.CrontabBinary).
		SetFs(h.fs)
	num, err := crontab.Remove()
	if err != nil {
		return err
	}
	if num == 0 {
		return ErrScheduledJobNotFound
	}
	return nil
}

// DisplayJobStatus has nothing to display (crond doesn't provide running information)
func (h *HandlerCrond) DisplayJobStatus(job *Config) error {
	return nil
}

func (h *HandlerCrond) Scheduled(profileName string) ([]Config, error) {
	crontab := crond.NewCrontab(nil).
		SetFile(h.config.CrontabFile).
		SetBinary(h.config.CrontabBinary).
		SetFs(h.fs)
	entries, configsErr := crontab.GetEntries()
	if configsErr != nil && len(entries) == 0 {
		return nil, configsErr
	}
	configs := make([]Config, 0, len(entries))
	for _, entry := range entries {
		profileName := entry.ProfileName()
		commandName := entry.CommandName()
		configFile := entry.ConfigFile()
		permission := constants.SchedulePermissionUser
		if entry.User() == "root" {
			permission = constants.SchedulePermissionSystem
		}

		if index := slices.IndexFunc(configs, func(cfg Config) bool {
			return cfg.ProfileName == profileName && cfg.CommandName == commandName && cfg.ConfigFile == configFile
		}); index >= 0 {
			configs[index].Schedules = append(configs[index].Schedules, entry.Event().String())
		} else {
			commandLine := entry.CommandLine()
			args := shell.SplitArguments(commandLine)
			configs = append(configs, Config{
				ProfileName:      profileName,
				CommandName:      commandName,
				ConfigFile:       configFile,
				Schedules:        []string{entry.Event().String()},
				Command:          args[0],
				Arguments:        NewCommandArguments(args[1:]),
				WorkingDirectory: entry.WorkDir(),
				Permission:       permission,
			})
		}
	}
	return configs, configsErr
}

// DetectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func (h *HandlerCrond) DetectSchedulePermission(p Permission) (Permission, bool) {
	switch p {
	case PermissionSystem, PermissionUserBackground, PermissionUserLoggedOn:
		// well defined
		return p, true

	default:
		// best guess is depending on the user being root or not:
		detected := PermissionUserBackground // sane default
		if os.Geteuid() == 0 {
			detected = PermissionSystem
		}
		// guess based on UID is never safe
		return detected, false
	}
}

// CheckPermission returns true if the user is allowed to access the job.
func (h *HandlerCrond) CheckPermission(user user.User, p Permission) bool {
	switch p {
	case PermissionUserLoggedOn, PermissionUserBackground:
		// user mode is always available
		return true

	default:
		if user.IsRoot() {
			return true
		}
		// last case is system (or undefined) + no sudo
		return false

	}
}

// init registers HandlerCrond
func init() {
	AddHandlerProvider(func(config SchedulerConfig, fallback bool) Handler {
		if config.Type() == constants.SchedulerCrond ||
			(fallback && config.Type() == constants.SchedulerOSDefault) {
			handler := NewHandlerCrond(config.Convert(constants.SchedulerCrond))
			return handler
		}
		return nil
	})
}
