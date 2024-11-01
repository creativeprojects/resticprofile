package schedule

import (
	"slices"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/crond"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
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
	return &HandlerCrond{config: cfg}
}

// Init verifies crond is available on this system
func (h *HandlerCrond) Init() error {
	if len(h.config.CrontabFile) > 0 {
		return nil
	}
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
func (h *HandlerCrond) CreateJob(job *Config, schedules []*calendar.Event, permission string) error {
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
	crontab := crond.NewCrontab(entries)
	crontab.SetFile(h.config.CrontabFile)
	crontab.SetBinary(h.config.CrontabBinary)
	err := crontab.Rewrite()
	if err != nil {
		return err
	}
	return nil
}

func (h *HandlerCrond) RemoveJob(job *Config, permission string) error {
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
	crontab := crond.NewCrontab(entries)
	crontab.SetFile(h.config.CrontabFile)
	crontab.SetBinary(h.config.CrontabBinary)
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
	crontab := crond.NewCrontab(nil)
	crontab.SetFile(h.config.CrontabFile)
	crontab.SetBinary(h.config.CrontabBinary)
	entries, err := crontab.GetEntries()
	if err != nil {
		return nil, err
	}
	configs := make([]Config, 0, len(entries))
	for _, entry := range entries {
		profileName := entry.ProfileName()
		commandName := entry.CommandName()
		configFile := entry.ConfigFile()
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
			})
		}
	}
	return configs, nil
}

// init registers HandlerCrond
func init() {
	AddHandlerProvider(func(config SchedulerConfig, fallback bool) Handler {
		if config.Type() == constants.SchedulerCrond ||
			(fallback && config.Type() == constants.SchedulerOSDefault) {
			handler := NewHandlerCrond(config.Convert(constants.SchedulerCrond))
			handler.fs = afero.NewOsFs()
			return handler
		}
		return nil
	})
}
