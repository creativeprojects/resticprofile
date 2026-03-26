package schedule

import (
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
)

// Config contains all information to schedule a profile command
type Config struct {
	ProfileName        string
	CommandName        string // restic command
	Schedules          []string
	Permission         string
	RunLevel           string
	WorkingDirectory   string
	Command            string // path to resticprofile executable
	Arguments          CommandArguments
	Environment        []string
	JobDescription     string
	TimerDescription   string
	Priority           string // Priority is either "background" or "standard"
	ConfigFile         string
	Flags              map[string]string // flags added to the command line
	AfterNetworkOnline bool
	SystemdDropInFiles []string
	HideWindow         bool
	StartWhenAvailable bool
	removeOnly         bool
}

// NewRemoveOnlyConfig creates a job config that may be used to call Job.Remove() on a scheduled job
func NewRemoveOnlyConfig(profileName, commandName string) *Config {
	return &Config{
		ProfileName: profileName,
		CommandName: commandName,
		removeOnly:  true,
	}
}

// SetCommand sets the command details for scheduling. Arguments are automatically
// processed to ensure proper handling of paths with spaces and special characters.
func (s *Config) SetCommand(wd, command string, args []string) {
	s.WorkingDirectory = wd
	s.Command = command
	s.Arguments = NewCommandArguments(args)
}

// GetPriority is either "background" or "standard"
func (s *Config) GetPriority() string {
	s.Priority = strings.ToLower(s.Priority)
	// default value for priority is "standard"
	if s.Priority != constants.SchedulePriorityBackground && s.Priority != constants.SchedulePriorityStandard {
		s.Priority = constants.SchedulePriorityStandard
	}
	return s.Priority
}

func (s *Config) GetFlag(name string) (string, bool) {
	if len(s.Flags) == 0 {
		return "", false
	}
	// we can't do a direct return, technically the map returns only one value
	value, found := s.Flags[name]
	return value, found
}

func (s *Config) SetFlag(name, value string) {
	if s.Flags == nil {
		s.Flags = make(map[string]string)
	}
	s.Flags[name] = value
}
