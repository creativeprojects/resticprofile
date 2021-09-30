package schedule

//
// Schedule: common code for all systems
//

// Scheduler
type Scheduler struct {
	handler Handler
}

// NewScheduler creates a Scheduler object
func NewScheduler(schedulerType SchedulerType, profileName string) *Scheduler {
	return &Scheduler{
		handler: NewHandler(schedulerType),
	}
}

// Init
func (s *Scheduler) Init() error {
	return s.handler.Init()
}

// Close
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
