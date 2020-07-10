package schedule

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
		return err
	}
	return nil
}
