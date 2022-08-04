package schedule

import (
	"fmt"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
)

// Handler interface for the scheduling software available on the system
type Handler interface {
	Init() error
	Close()
	ParseSchedules(schedules []string) ([]*calendar.Event, error)
	DisplayParsedSchedules(command string, events []*calendar.Event)
	DisplaySchedules(command string, schedules []string) error
	DisplayStatus(profileName string) error
	CreateJob(job *config.ScheduleConfig, schedules []*calendar.Event, permission string) error
	RemoveJob(job *config.ScheduleConfig, permission string) error
	DisplayJobStatus(job *config.ScheduleConfig) error
}

func lookupBinary(name, binary string) error {
	found, err := exec.LookPath(binary)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system (cannot find %q command in path)", name, binary)
	}
	return nil
}
