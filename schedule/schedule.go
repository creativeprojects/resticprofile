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
)

// Scheduler interface
type Scheduler interface {
	Init() error
	Close()
	NewJob(Config) SchedulerJob
	DisplayStatus()
}
