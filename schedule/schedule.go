package schedule

//
// Schedule: common code for all systems
//

// Scheduler
type Scheduler struct {
	handler Handler
}

// NewScheduler creates a Scheduler object
func NewScheduler(scheduler SchedulerType, profileName string) *Scheduler {
	return &Scheduler{
		handler: NewHandler(),
	}
}

// Init verifies launchd is available on this system
func (s *Scheduler) Init() error {
	return s.handler.Init()
}

// Close does nothing with launchd
func (s *Scheduler) Close() {
	s.handler.Close()
}

// NewJob instantiates a Job object (of SchedulerJob interface) to schedule jobs
func (s *Scheduler) NewJob(config Config) SchedulerJob {
	return &Job{
		config:  config,
		handler: s.handler,
	}
}

// DisplayStatus does nothing on launchd
func (s *Scheduler) DisplayStatus() {
}
