package config

import (
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util"
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

// Schedule is an intermediary object between the configuration (v1, v2+) and the ScheduleConfig object used by the scheduler.
// The object is also used to display the scheduling configuration
type Schedule struct {
	CommandName             string            `mapstructure:"run"`
	Group                   string            `mapstructure:"group"`    // v2+ only
	Profiles                []string          `mapstructure:"profiles"` // multiple profiles in v2+ only
	Schedules               []string          `mapstructure:"schedule"`
	Permission              string            `mapstructure:"permission"`
	Log                     string            `mapstructure:"log"`
	Priority                string            `mapstructure:"priority"`
	LockMode                string            `mapstructure:"lock-mode"`
	LockWait                time.Duration     `mapstructure:"lock-wait"`
	Environment             []string          `mapstructure:"environment"`
	IgnoreOnBattery         bool              `mapstructure:"ignore-on-battery"`
	IgnoreOnBatteryLessThan int               `mapstructure:"ignore-on-battery-less-than"`
	AfterNetworkOnline      bool              `mapstructure:"after-network-online"`
	SystemdDropInFiles      []string          `mapstructure:"systemd-drop-in-files"`
	ConfigFile              string            `show:"noshow"`
	Flags                   map[string]string `show:"noshow"`
}

func NewEmptySchedule(profileName, command string) *Schedule {
	return &Schedule{
		Profiles:    []string{profileName},
		CommandName: command,
	}
}

func (s *Schedule) Init(config *Config, profiles ...*Profile) {
	// populate profiles from group (v2+ only)

	// temporary log file
	if s.Log != "" {
		if tempDir, err := util.TempDir(); err == nil && strings.HasPrefix(s.Log, filepath.ToSlash(tempDir)) {
			s.Log = path.Join(constants.TemporaryDirMarker, s.Log[len(tempDir):])
		}
	}
}

func (s *Schedule) GetLockMode() ScheduleLockMode {
	switch s.LockMode {
	case constants.ScheduleLockModeOptionFail:
		return ScheduleLockModeFail
	case constants.ScheduleLockModeOptionIgnore:
		return ScheduleLockModeIgnore
	default:
		return ScheduleLockModeDefault
	}
}

func (s *Schedule) GetLockWait() time.Duration {
	if s.LockWait <= 2*time.Second {
		return 0
	}
	return s.LockWait
}

func (s *Schedule) GetFlag(name string) (string, bool) {
	if len(s.Flags) == 0 {
		return "", false
	}
	// we can't do a direct return, technically the map returns only one value
	value, found := s.Flags[name]
	return value, found
}

func (s *Schedule) SetFlag(name, value string) {
	if s.Flags == nil {
		s.Flags = make(map[string]string)
	}
	s.Flags[name] = value
}
