//go:build !darwin && !windows

package schedule

import (
	"io"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/crond"
)

const (
	crontabBinary = "crontab"
)

// HandlerCrond is a handler for crond scheduling
type HandlerCrond struct {
	config SchedulerConfig
}

// NewHandlerCrond creates a new handler for crond scheduling
func NewHandlerCrond(config SchedulerConfig) *HandlerCrond {
	return &HandlerCrond{
		config: config,
	}
}

// Init verifies crond is available on this system
func (h *HandlerCrond) Init() error {
	return lookupBinary("crond", crontabBinary)
}

// Close does nothing with crond
func (h *HandlerCrond) Close() {}

func (h *HandlerCrond) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

func (h *HandlerCrond) DisplayParsedSchedules(command string, events []*calendar.Event) {
	displayParsedSchedules(command, events)
}

// DisplaySchedules does nothing with launchd
func (h *HandlerCrond) DisplaySchedules(command string, schedules []string) error {
	return nil
}

// DisplayStatus does nothing with launchd
func (h *HandlerCrond) DisplayStatus(profileName string, w io.Writer) error {
	return nil
}

// CreateJob is creating the crontab
func (h *HandlerCrond) CreateJob(job JobConfig, schedules []*calendar.Event) error {
	permission := getSchedulePermission(job.Permission())
	ok := checkPermission(permission)
	if !ok {
		return permissionError("create")
	}
	entries := make([]crond.Entry, len(schedules))
	for i, event := range schedules {
		entries[i] = crond.NewEntry(
			event,
			job.Configfile(),
			job.Title(),
			job.SubTitle(),
			job.Command()+" "+strings.Join(job.Arguments(), " "),
			job.WorkingDirectory(),
		)
	}
	crontab := crond.NewCrontab(entries)
	err := crontab.Rewrite()
	if err != nil {
		return err
	}
	return nil
}

func (h *HandlerCrond) RemoveJob(job JobConfig) error {
	removeOnly := isRemoveOnlyConfig(job)
	var permission string
	if removeOnly {
		permission, _ = detectSchedulePermission(job.Permission()) // silent call for possibly non-existent job
	} else {
		permission = getSchedulePermission(job.Permission())
	}

	ok := checkPermission(permission)
	if !ok {
		return permissionError("remove")
	}
	entries := []crond.Entry{
		crond.NewEntry(
			calendar.NewEvent(),
			job.Configfile(),
			job.Title(),
			job.SubTitle(),
			job.Command()+" "+strings.Join(job.Arguments(), " "),
			job.WorkingDirectory(),
		),
	}
	crontab := crond.NewCrontab(entries)
	err := crontab.Remove()
	if err != nil {
		return err
	}
	return nil
}

// DisplayJobStatus has nothing to display (crond doesn't provide running information)
func (h *HandlerCrond) DisplayJobStatus(job JobConfig, w io.Writer) error {
	return nil
}

var (
	_ Handler = &HandlerCrond{}
)
