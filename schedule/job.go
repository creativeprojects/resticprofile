package schedule

import "errors"

//
// Schedule: common code for all systems
//

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
	config    Config
	scheduler SchedulerType
}

var ErrorJobCanBeRemovedOnly = errors.New("job can be removed only")

// Accessible checks if the current user is permitted to access the job
func (j *Job) Accessible() bool {
	permission, _ := j.detectSchedulePermission()
	return j.checkPermission(permission)
}

// Create a new job
func (j *Job) Create() error {
	if j.RemoveOnly() {
		return ErrorJobCanBeRemovedOnly
	}

	schedules, err := j.loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.createJob(schedules)
	if err != nil {
		return err
	}

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

// RemoveOnly returns true if this job can be removed only
func (j *Job) RemoveOnly() bool {
	return isRemoveOnlyConfig(j.config)
}

// Status of a job
func (j *Job) Status() error {
	if j.RemoveOnly() {
		return ErrorJobCanBeRemovedOnly
	}

	_, err := j.loadSchedules(j.config.SubTitle(), j.config.Schedules())
	if err != nil {
		return err
	}

	err = j.displayStatus(j.config.SubTitle())
	if err != nil {
		return err
	}
	return nil
}

// Verify interface
var _ SchedulerJob = &Job{}
