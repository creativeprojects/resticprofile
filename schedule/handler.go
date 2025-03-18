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
	DisplaySchedules(profile, command string, schedules []string) error
	DisplayStatus(profileName string) error
	CreateJob(job *Config, schedules []*calendar.Event, permission Permission) error
	RemoveJob(job *Config, permission Permission) error
	DisplayJobStatus(job *Config) error
	// Scheduled can return configs at the same time as an error: it means some configs are fine but some others cannot be loaded properly
	Scheduled(profileName string) ([]Config, error)
	// DetectSchedulePermission returns the permission defined from the configuration,
	// or the best guess considering the current user permission.
	// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
	DetectSchedulePermission(permission Permission) (Permission, bool)
	// CheckPermission returns true if the user is allowed to access the job.
	CheckPermission(p Permission) bool
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
		return fmt.Errorf("cannot find %q: please ensure %s is installed and available on this system", binary, name)
	}
	return nil
}
