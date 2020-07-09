package schedule

import (
	"fmt"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
)

// Job scheduler
type Job struct {
	configFile string
	profile    *config.Profile
	schedules  map[config.ScheduledCommand][]*calendar.Event
}

// NewJob instantiates a Job object to schedule jobs
func NewJob(configFile string, profile *config.Profile) *Job {
	return &Job{
		configFile: configFile,
		profile:    profile,
	}
}

// Create a new job
func (j *Job) Create() error {
	err := checkSystem()
	if err != nil {
		return err
	}

	err = j.checkSchedules()
	if err != nil {
		return err
	}

	for command, schedules := range j.schedules {
		err = j.createJob(command.String(), schedules)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update an existing job
func (j *Job) Update() error {
	err := checkSystem()
	if err != nil {
		return err
	}
	return nil
}

// Remove a job
func (j *Job) Remove() error {
	err := checkSystem()
	if err != nil {
		return err
	}
	err = j.removeJob()
	if err != nil {
		return err
	}
	return nil
}

// Status of a job
func (j *Job) Status() error {
	err := checkSystem()
	if err != nil {
		return err
	}

	err = j.checkSchedules()
	if err != nil {
		return err
	}

	for command := range j.schedules {
		err = j.displayStatus(command.String())
		if err != nil {
			return err
		}
	}
	return nil
}

// checkSchedules for each command and load the schedules into j.schedules
func (j *Job) checkSchedules() error {
	var err error
	j.schedules = make(map[config.ScheduledCommand][]*calendar.Event, 3)
	commandSchedules := j.profile.GetScheduledCommands()
	if len(commandSchedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", j.profile.Name)
	}
	for command, schedules := range commandSchedules {
		j.schedules[command], err = loadSchedules(command.String(), schedules.Schedule)
		if err != nil {
			return err
		}
	}
	return nil
}
