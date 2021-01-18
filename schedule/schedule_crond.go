//+build !darwin,!windows

package schedule

import (
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/crond"
)

const (
	crontabBin = "crontab"
)

// createCrondJob is creating the crontab
func (j *Job) createCrondJob(schedules []*calendar.Event) error {
	entries := make([]crond.Entry, len(schedules))
	for i, event := range schedules {
		entries[i] = crond.NewEntry(event, j.config.Configfile(), j.config.Title(), j.config.SubTitle(), j.config.Command()+" "+strings.Join(j.config.Arguments(), " "))
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
		crond.NewEntry(calendar.NewEvent(), j.config.Configfile(), j.config.Title(), j.config.SubTitle(), j.config.Command()+" "+strings.Join(j.config.Arguments(), " ")),
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
