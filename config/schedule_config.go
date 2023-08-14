package config

import (
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
)

type ScheduleLockMode int8

const (
	// ScheduleLockModeDefault waits on acquiring a lock (local and repository) for up to ScheduleConfig lockWait (duration), before failing a schedule.
	// With lockWait set to 0, ScheduleLockModeDefault and ScheduleLockModeFail behave the same.
	ScheduleLockModeDefault = ScheduleLockMode(0)
	// ScheduleLockModeFail fails immediately on a lock failure without waiting.
	ScheduleLockModeFail = ScheduleLockMode(1)
	// ScheduleLockModeIgnore does not create or fail on resticprofile locks. Repository locks cause an immediate failure.
	ScheduleLockModeIgnore = ScheduleLockMode(2)
)

// ScheduleConfig contains all information to schedule a profile command
type ScheduleConfig struct {
	Title                   string
	SubTitle                string
	Schedules               []string
	Permission              string
	WorkingDirectory        string
	Command                 string
	Arguments               []string
	Environment             []string
	JobDescription          string
	TimerDescription        string
	Priority                string
	Log                     string
	LockMode                string
	LockWait                time.Duration
	ConfigFile              string
	Flags                   map[string]string
	RemoveOnly              bool
	IgnoreOnBattery         bool
	IgnoreOnBatteryLessThan int
}

// NewRemoveOnlyConfig creates a job config that may be used to call Job.Remove() on a scheduled job
func NewRemoveOnlyConfig(profileName, commandName string) *ScheduleConfig {
	return &ScheduleConfig{
		Title:      profileName,
		SubTitle:   commandName,
		RemoveOnly: true,
	}
}

func (s *ScheduleConfig) SetCommand(wd, command string, args []string) {
	s.WorkingDirectory = wd
	s.Command = command
	s.Arguments = args
}

// Priority is either "background" or "standard"
func (s *ScheduleConfig) GetPriority() string {
	s.Priority = strings.ToLower(s.Priority)
	// default value for priority is "background"
	if s.Priority != constants.SchedulePriorityBackground && s.Priority != constants.SchedulePriorityStandard {
		s.Priority = constants.SchedulePriorityBackground
	}
	return s.Priority
}

func (s *ScheduleConfig) GetLockMode() ScheduleLockMode {
	switch s.LockMode {
	case constants.ScheduleLockModeOptionFail:
		return ScheduleLockModeFail
	case constants.ScheduleLockModeOptionIgnore:
		return ScheduleLockModeIgnore
	default:
		return ScheduleLockModeDefault
	}
}

func (s *ScheduleConfig) GetLockWait() time.Duration {
	if s.LockWait <= 2*time.Second {
		return 0
	}
	return s.LockWait
}

func (s *ScheduleConfig) GetFlag(name string) (string, bool) {
	if len(s.Flags) == 0 {
		return "", false
	}
	// we can't do a direct return, technically the map returns only one value
	value, found := s.Flags[name]
	return value, found
}

func (s *ScheduleConfig) SetFlag(name, value string) {
	if s.Flags == nil {
		s.Flags = make(map[string]string)
	}
	s.Flags[name] = value
}

func (s *ScheduleConfig) Export() Schedule {
	return Schedule{
		Profiles:   []string{s.Title},
		Command:    s.SubTitle,
		Permission: s.Permission,
		Log:        s.Log,
		Priority:   s.Priority,
		LockMode:   s.LockMode,
		LockWait:   s.LockWait,
		Schedule:   s.Schedules,
	}
}
