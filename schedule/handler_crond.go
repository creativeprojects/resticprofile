package schedule

import (
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/crond"
	"github.com/creativeprojects/resticprofile/platform"
)

var crontabBinary = "crontab"

// HandlerCrond is a handler for crond scheduling
type HandlerCrond struct {
	config SchedulerCrond
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

func (h *HandlerCrond) DisplayParsedSchedules(command string, events []*calendar.Event) {
	displayParsedSchedules(command, events)
}

// DisplaySchedules does nothing with crond
func (h *HandlerCrond) DisplaySchedules(command string, schedules []string) error {
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
			job.Command+" "+strings.Join(job.Arguments, " "),
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
			job.Command+" "+strings.Join(job.Arguments, " "),
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
		return ErrServiceNotFound
	}
	return nil
}

// DisplayJobStatus has nothing to display (crond doesn't provide running information)
func (h *HandlerCrond) DisplayJobStatus(job *Config) error {
	return nil
}

// init registers HandlerCrond
func init() {
	AddHandlerProvider(func(config SchedulerConfig, fallback bool) (hr Handler) {
		if config.Type() == constants.SchedulerCrond ||
			(fallback && config.Type() == constants.SchedulerOSDefault) {
			hr = NewHandlerCrond(config.Convert(constants.SchedulerCrond))
		}
		return
	})
}
