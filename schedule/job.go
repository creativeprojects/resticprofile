package schedule

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
}

// SchedulerJob interface
type SchedulerJob interface {
	Create() error
	Update() error
	Remove() error
	Status() error
}

// Job scheduler
type Job struct {
	config Config
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

// Verify interface
var _ SchedulerJob = &Job{}
