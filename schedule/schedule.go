package schedule

//
// Schedule: common code for all systems
//

import (
	"github.com/creativeprojects/resticprofile/constants"
)

var (
	// ScheduledSections are the command that can be scheduled (backup, retention, check, prune)
	ScheduledSections = []string{
		constants.CommandBackup,
		constants.SectionConfigurationRetention,
		constants.CommandCheck,
		constants.CommandForget,
		constants.CommandPrune,
	}
	// Scheduler is the scheduler to use on this system
	Scheduler = ""
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
	Priority() string
	Logfile() string
	Configfile() string
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
	return nil
}

// Remove a job
func (j *Job) Remove() error {
	err := j.removeJob()
	if err != nil {
		return err
	}
	return nil
}

// Status of a job
func (j *Job) Status() error {
	_, err := loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.displayStatus(j.config.SubTitle())
	if err != nil {
		return err
	}
	return nil
}
