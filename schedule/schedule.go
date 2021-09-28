package schedule

//
// Schedule: common code for all systems
//

// Scheduler interface
type Scheduler interface {
	Init() error
	Close()
	NewJob(Config) SchedulerJob
	DisplayStatus()
}
