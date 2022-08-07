package schedule

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
)

// Scheduler
type Scheduler struct {
	profileName string
	handler     Handler
}

// NewScheduler creates a Scheduler object
func NewScheduler(handler Handler, profileName string) *Scheduler {
	return &Scheduler{
		profileName: profileName,
		handler:     handler,
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
func (s *Scheduler) NewJob(config *config.ScheduleConfig) SchedulerJob {
	return &Job{
		config:  config,
		handler: s.handler,
	}
}

// DisplayStatus
func (s *Scheduler) DisplayStatus() {
	err := s.handler.DisplayStatus(s.profileName)
	if err != nil {
		clog.Error(err)
	}
}
