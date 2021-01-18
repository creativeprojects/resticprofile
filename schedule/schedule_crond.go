//+build !darwin,!windows

package schedule

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/crond"
)

const (
	crontabBin = "crontab"
)

// CrondSchedule is a Scheduler using crond
type CrondSchedule struct {
}

// Init verifies crond is available on this system
func (s *CrondSchedule) Init() error {
	found, err := exec.LookPath(crontabBin)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like crond is installed on your system (cannot find %q command in path)", crontabBin)
	}
	return nil
}

// Close does nothing when using crond
func (s *CrondSchedule) Close() {
}

// NewJob instantiates a Job object (of SchedulerJob interface) to schedule jobs
func (s *CrondSchedule) NewJob(config Config) SchedulerJob {
	return &Job{
		config:    config,
		scheduler: constants.SchedulerCrond,
	}
}

// Verify interface
var _ Scheduler = &CrondSchedule{}

// createCrondJob is creating the crontab
func (j *Job) createCrondJob(schedules []*calendar.Event) error {
	entries := make([]crond.Entry, len(schedules))
	for i, event := range schedules {
		entries[i] = crond.NewEntry(
			event,
			j.config.Configfile(),
			j.config.Title(),
			j.config.SubTitle(),
			j.config.Command()+" "+strings.Join(j.config.Arguments(), " "),
			j.config.WorkingDirectory(),
		)
	}
	crontab := crond.NewCrontab(entries)
	err := crontab.Rewrite()
	if err != nil {
		return err
	}
	return nil
}

func (j *Job) removeCrondJob() error {
	entries := []crond.Entry{
		crond.NewEntry(
			calendar.NewEvent(),
			j.config.Configfile(),
			j.config.Title(),
			j.config.SubTitle(),
			j.config.Command()+" "+strings.Join(j.config.Arguments(), " "),
			j.config.WorkingDirectory(),
		),
	}
	crontab := crond.NewCrontab(entries)
	err := crontab.Remove()
	if err != nil {
		return err
	}
	return nil
}

// displayCrondStatus has nothing to display (crond doesn't provide running information)
func (j *Job) displayCrondStatus(command string) error {
	return nil
}
