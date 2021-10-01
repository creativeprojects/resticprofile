package schedule

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
)

type Handler interface {
	Init() error
	Close()
	ParseSchedules(schedules []string) ([]*calendar.Event, error)
	DisplayParsedSchedules(command string, events []*calendar.Event)
	DisplaySchedules(command string, schedules []string) error
	DisplayStatus(profileName string, w io.Writer) error
	CreateJob(job JobConfig, schedules []*calendar.Event) error
	RemoveJob(job JobConfig) error
	DisplayJobStatus(job JobConfig, w io.Writer) error
}

func lookupBinary(name, binary string) error {
	found, err := exec.LookPath(binary)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system (cannot find %q command in path)", name, binary)
	}
	return nil
}
