package schedule

import (
	"fmt"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
)

// Handler interface for the scheduling software available on the system
type Handler interface {
	Init() error
	Close()
	ParseSchedules(schedules []string) ([]*calendar.Event, error)
	DisplayParsedSchedules(command string, events []*calendar.Event)
	DisplaySchedules(command string, schedules []string) error
	DisplayStatus(profileName string) error
	CreateJob(job *Config, schedules []*calendar.Event, permission string) error
	RemoveJob(job *Config, permission string) error
	DisplayJobStatus(job *Config) error
}

// FindHandler creates a schedule handler depending on the configuration or nil if the config is not supported
func FindHandler(config SchedulerConfig) Handler {
	for _, fallback := range []bool{false, true} {
		for _, provider := range providers {
			if handler := provider(config, fallback); handler != nil {
				return handler
			}
		}
	}
	return nil
}

// NewHandler creates a schedule handler depending on the configuration, panics if the config is not supported
func NewHandler(config SchedulerConfig) Handler {
	if h := FindHandler(config); h != nil {
		return h
	}
	panic(fmt.Errorf("scheduler %q is not supported in this environment", config.Type()))
}

type HandlerProvider func(config SchedulerConfig, fallback bool) Handler

var providers []HandlerProvider

// AddHandlerProvider allows to register a provider for a SchedulerConfig handler
func AddHandlerProvider(provider HandlerProvider) {
	providers = append(providers, provider)
}

func lookupBinary(name, binary string) error {
	found, err := exec.LookPath(binary)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system (cannot find %q in path)", name, binary)
	}
	return nil
}
