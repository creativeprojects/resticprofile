package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/term"
)

//
// Job: common code for all systems
//

// JobConfig contains all the information needed to schedule a Job
type JobConfig interface {
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
	GetFlag(string) (string, bool)
}

// SchedulerJob interface
type SchedulerJob interface {
	Accessible() bool
	Create() error
	Remove() error
	RemoveOnly() bool
	Status() error
}

// Job scheduler
type Job struct {
	config  JobConfig
	handler Handler
}

var ErrorJobCanBeRemovedOnly = errors.New("job can be removed only")

// Accessible checks if the current user is permitted to access the job
func (j *Job) Accessible() bool {
	permission, _ := detectSchedulePermission(j.config.Permission())
	return checkPermission(permission)
}

// Create a new job
func (j *Job) Create() error {
	if j.RemoveOnly() {
		return ErrorJobCanBeRemovedOnly
	}

	schedules, err := j.handler.ParseSchedules(j.config.Schedules())
	if err != nil {
		return err
	}

	if len(schedules) > 0 {
		j.handler.DisplayParsedSchedules(j.config.SubTitle(), schedules)
	} else {
		err := j.handler.DisplaySchedules(j.config.SubTitle(), j.config.Schedules())
		if err != nil {
			return err
		}
	}

	err = j.handler.CreateJob(j.config, schedules)
	if err != nil {
		return err
	}

	return nil
}

// Remove a job
func (j *Job) Remove() error {
	err := j.handler.RemoveJob(j.config)
	if err != nil {
		return err
	}
	return nil
}

// RemoveOnly returns true if this job can be removed only
func (j *Job) RemoveOnly() bool {
	return isRemoveOnlyConfig(j.config)
}

// Status of a job
func (j *Job) Status() error {
	if j.RemoveOnly() {
		return ErrorJobCanBeRemovedOnly
	}

	schedules, err := j.handler.ParseSchedules(j.config.Schedules())
	if err != nil {
		return err
	}

	if len(schedules) > 0 {
		j.handler.DisplayParsedSchedules(j.config.SubTitle(), schedules)
	} else {
		err := j.handler.DisplaySchedules(j.config.SubTitle(), j.config.Schedules())
		if err != nil {
			return err
		}
	}

	err = j.handler.DisplayJobStatus(j.config, term.GetOutput())
	if err != nil {
		return err
	}
	return nil
}

// Verify interface
var _ SchedulerJob = &Job{}
