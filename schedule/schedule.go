package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

var (
	// ScheduledSections are the command that can be scheduled (backup, retention, check)
	ScheduledSections = []string{
		constants.CommandBackup,
		constants.SectionConfigurationRetention,
		constants.CommandCheck,
	}
)

// Config contains all the information needed to schedule a Job
type Config interface {
	Title() string
	SubTitle() string
	JobDescription() string
	TimerDescription() string
	Schedules() []string
	Permission() string
	WorkingDirectory() string
	Command() string
	Arguments() []string
	Environment() map[string]string
}

// Job scheduler
type Job struct {
	config Config
}

// NewJob instantiates a Job object to schedule jobs
func NewJob(config Config) *Job {
	return &Job{
		config: config,
	}
}

// Create a new job
func (j *Job) Create() error {
	err := checkSystem()
	if err != nil {
		return err
	}

	schedules, err := loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.createJob(schedules)
	if err != nil {
		return err
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

	_, err = loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.displayStatus(j.config.SubTitle())
	if err != nil {
		if errors.Is(err, ErrorServiceNotFound) {
			// Display a warning and keep going
			clog.Warningf("service %s/%s not found", j.config.Title(), j.config.SubTitle())
			return nil
		}
		return err
	}
	return nil
}
