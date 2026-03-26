package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/spf13/cast"
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

// ScheduleBaseConfig is the base user configuration that could be shared across all schedules.
type ScheduleBaseConfig struct {
	Permission              string         `mapstructure:"permission" default:"auto" enum:"auto;system;user;user_logged_on" description:"Specify whether the schedule runs with system or user privileges - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	RunLevel                string         `mapstructure:"run-level" default:"auto" enum:"auto;lowest;highest" description:"Specify the schedule privilege level (for Windows Task Scheduler only)"`
	Log                     string         `mapstructure:"log" examples:"/resticprofile.log;syslog-tcp://syslog-server:514;syslog:server;syslog:" description:"Redirect the output into a log file or to syslog when running on schedule - see https://creativeprojects.github.io/resticprofile/configuration/logs/"`
	CommandOutput           string         `mapstructure:"command-output" default:"auto" enum:"auto;log;console;all" description:"Sets the destination for command output (stderr/stdout). \"log\" sends output to the log file (if specified), \"console\" sends it to the console instead. \"auto\" sends it to \"both\" if console is a terminal otherwise to \"log\" only - see https://creativeprojects.github.io/resticprofile/configuration/logs/"`
	Priority                string         `mapstructure:"priority" default:"standard" enum:"background;standard" description:"Set the priority at which the schedule is run"`
	LockMode                string         `mapstructure:"lock-mode" default:"default" enum:"default;fail;ignore" description:"Specify how locks are used when running on schedule - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	LockWait                maybe.Duration `mapstructure:"lock-wait" examples:"150s;15m;30m;45m;1h;2h30m" description:"Set the maximum time to wait for acquiring locks when running on schedule"`
	EnvCapture              []string       `mapstructure:"capture-environment" default:"RESTIC_*" description:"Set names (or glob expressions) of environment variables to capture during schedule creation. The captured environment is applied prior to \"profile.env\" when running the schedule. Whether capturing is supported depends on the type of scheduler being used (supported in \"systemd\" and \"launchd\")"`
	IgnoreOnBattery         maybe.Bool     `mapstructure:"ignore-on-battery" default:"false" description:"Don't start this schedule when running on battery"`
	IgnoreOnBatteryLessThan int            `mapstructure:"ignore-on-battery-less-than" default:"" examples:"20;33;50;75" description:"Don't start this schedule when running on battery and the state of charge is less than this percentage"`
	AfterNetworkOnline      maybe.Bool     `mapstructure:"after-network-online" description:"Don't start this schedule when the network is offline (supported in \"systemd\")"`
	SystemdDropInFiles      []string       `mapstructure:"systemd-drop-in-files" default:"" description:"Files containing systemd drop-in (override) files - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	HideWindow              maybe.Bool     `mapstructure:"hide-window" default:"false" description:"Hide schedule window when running in foreground (Windows only)"`
	StartWhenAvailable      maybe.Bool     `mapstructure:"start-when-available" default:"false" description:"Start the task as soon as possible after a scheduled start is missed (Windows only)"`
}

// scheduleBaseConfigDefaults declares built-in scheduling defaults
var scheduleBaseConfigDefaults = ScheduleBaseConfig{
	Permission:    "auto",
	RunLevel:      "auto",
	CommandOutput: constants.DefaultCommandOutput,
	Priority:      "standard",
	LockMode:      "default",
	EnvCapture:    []string{"RESTIC_*"},
}

func (s *ScheduleBaseConfig) init(defaults *ScheduleBaseConfig) {
	// defaults
	if defaults == nil {
		defaults = &scheduleBaseConfigDefaults
	}
	if s.Permission == "" {
		s.Permission = defaults.Permission
	}
	if s.RunLevel == "" {
		s.RunLevel = defaults.RunLevel
	}
	if s.Log == "" {
		s.Log = defaults.Log
	}
	if s.CommandOutput == "" {
		s.CommandOutput = defaults.CommandOutput
	}
	if s.Priority == "" {
		s.Priority = defaults.Priority
	}
	if s.LockMode == "" {
		s.LockMode = defaults.LockMode
	}
	if !s.LockWait.HasValue() {
		s.LockWait = defaults.LockWait
	}
	if s.EnvCapture == nil {
		s.EnvCapture = slices.Clone(defaults.EnvCapture)
	}
	if !s.IgnoreOnBattery.HasValue() {
		s.IgnoreOnBattery = defaults.IgnoreOnBattery
	}
	if !s.AfterNetworkOnline.HasValue() {
		s.AfterNetworkOnline = defaults.AfterNetworkOnline
	}
	if s.SystemdDropInFiles == nil {
		s.SystemdDropInFiles = slices.Clone(defaults.SystemdDropInFiles)
	}
	if !s.HideWindow.HasValue() {
		s.HideWindow = defaults.HideWindow
	}
	if !s.StartWhenAvailable.HasValue() {
		s.StartWhenAvailable = defaults.StartWhenAvailable
	}
}

func (s *ScheduleBaseConfig) applyOverrides(section *ScheduleBaseSection) {
	// capture a copy of self as defaults
	defaults := *s
	// applying the settings of the section
	s.Permission = section.SchedulePermission
	s.RunLevel = section.ScheduleRunLevel
	s.Log = section.ScheduleLog
	s.Priority = section.SchedulePriority
	s.LockMode = section.ScheduleLockMode
	s.LockWait = section.ScheduleLockWait
	s.EnvCapture = slices.Clone(section.ScheduleEnvCapture)
	s.IgnoreOnBattery = section.ScheduleIgnoreOnBattery
	s.AfterNetworkOnline = section.ScheduleAfterNetworkOnline
	s.HideWindow = section.ScheduleHideWindow
	s.StartWhenAvailable = section.ScheduleStartWhenAvailable
	// re-init with defaults
	s.init(&defaults)
}

func (s *ScheduleBaseConfig) applyProfile(profile *Profile) {
	// capture a copy of self as defaults
	defaults := *s
	// applying the settings from the profile
	s.SystemdDropInFiles = slices.Clone(profile.SystemdDropInFiles)
	// re-init with defaults
	s.init(&defaults)
}

type ScheduleOriginType int

const (
	ScheduleOriginProfile ScheduleOriginType = iota
	ScheduleOriginGroup
)

type ScheduleConfigOrigin struct {
	Type          ScheduleOriginType
	Name, Command string
}

func (o ScheduleConfigOrigin) Compare(other ScheduleConfigOrigin) (c int) {
	c = int(other.Type) - int(o.Type) // groups first
	if c == 0 {
		c = strings.Compare(o.Name, other.Name)
	}
	if c == 0 {
		c = strings.Compare(o.Command, other.Command)
	}
	return
}

func (o ScheduleConfigOrigin) String() string {
	return fmt.Sprintf("%s@%s", o.Command, o.Name)
}

// ScheduleOrigin returns a origin for the specified name command and optional type (defaulting to ScheduleOriginProfile)
func ScheduleOrigin(name, command string, kind ...ScheduleOriginType) (s ScheduleConfigOrigin) {
	s.Name = name
	s.Command = command
	if len(kind) == 1 {
		s.Type = kind[0]
	}
	return
}

// ScheduleConfig is the user configuration of a specific schedule bound to a command in a profile or group.
type ScheduleConfig struct {
	normalized bool
	origin     ScheduleConfigOrigin `show:"noshow"`
	Schedules  []string             `mapstructure:"at" examples:"hourly;daily;weekly;monthly;10:00,14:00,18:00,22:00;Wed,Fri 17:48;*-*-15 02:45;Mon..Fri 00:30" description:"Set the times at which the scheduled command is run (times are specified in systemd timer format)"`

	ScheduleBaseConfig `mapstructure:",squash"`
}

// NewDefaultScheduleConfig returns a new schedule configuration that is initialized with defaults
func NewDefaultScheduleConfig(config *Config, origin ScheduleConfigOrigin, schedules ...string) (s *ScheduleConfig) {
	var defaults *ScheduleBaseConfig
	if config != nil {
		defaults = config.mustGetGlobalSection().ScheduleDefaults
	}

	s = new(ScheduleConfig)
	if len(schedules) > 0 {
		s.setSchedules(schedules)
	}
	s.init(defaults)
	s.origin = origin
	return s
}

func newScheduleConfig(profile *Profile, section *ScheduleBaseSection) (s *ScheduleConfig) {
	origin := ScheduleConfigOrigin{} // is set later
	config := profile.config

	// decode ScheduleBaseSection.Schedule
	switch expression := section.Schedule.(type) {
	case string:
		s = NewDefaultScheduleConfig(config, origin, expression)
		s.applyProfile(profile)
	case []string, []any:
		s = NewDefaultScheduleConfig(config, origin, cast.ToStringSlice(expression)...)
		s.applyProfile(profile)
	default:
		if expression != nil {
			cfg := new(ScheduleConfig)
			decoder, err := config.newUnmarshaller(cfg)
			if err == nil {
				err = decoder.Decode(expression)
			}
			if err == nil {
				defaults := NewDefaultScheduleConfig(config, origin) // applying defaults after parsing to avoid side effects
				defaults.applyProfile(profile)
				cfg.init(&defaults.ScheduleBaseConfig)
				s = cfg
			} else {
				if bytes, e := json.Marshal(expression); e == nil {
					expression = string(bytes)
				}
				clog.Errorf("failed decoding schedule %v: %s", expression, err.Error())
			}
		}
	}

	// init
	if s.HasSchedules() {
		s.applyOverrides(section)
	} else {
		s = nil
	}
	return
}

func (s *ScheduleConfig) setSchedules(schedules []string) {
	schedules = collect.From(schedules, strings.TrimSpace)
	schedules = collect.All(schedules, func(at string) bool { return len(at) > 0 })
	s.Schedules = schedules
	s.normalized = true
}

// HasSchedules returns true if the normalized list of schedules is not empty.
// The func is nil tolerant and returns false for config.Schedule(nil).HasSchedules()
func (s *ScheduleConfig) HasSchedules() bool {
	if s == nil {
		return false
	}
	if !s.normalized {
		s.setSchedules(s.Schedules)
	}
	return len(s.Schedules) > 0
}

func (s *ScheduleConfig) ScheduleOrigin() ScheduleConfigOrigin {
	return s.origin
}

// Schedulable may be implemented by sections that can provide command schedules (= groups and profiles)
type Schedulable interface {
	// Schedules returns a command to schedule map
	Schedules() map[string]*Schedule
	// SchedulableCommands returns a list of commands that can be scheduled
	SchedulableCommands() []string
	// Kind returns the kind of the schedule origin (profile or group)
	Kind() string
}

// Schedule is the configuration used in profiles and groups for passing the user config to the scheduler system.
type Schedule struct {
	ScheduleConfig

	ConfigFile  string            `show:"noshow"`
	Environment []string          `show:"noshow"`
	Flags       map[string]string `show:"noshow"`
}

// NewDefaultSchedule creates a new Schedule for the specified ScheduleConfigOrigin that is initialized with defaults
func NewDefaultSchedule(config *Config, origin ScheduleConfigOrigin, schedules ...string) *Schedule {
	return NewSchedule(config, NewDefaultScheduleConfig(config, origin, schedules...))
}

// NewSchedule creates a new Schedule for the specified Config and ScheduleConfig
func NewSchedule(config *Config, sc *ScheduleConfig) *Schedule {
	return newSchedule(config, sc, nil)
}

// newScheduleForProfile creates a Schedule for the given Profile and ScheduleConfig
func newScheduleForProfile(profile *Profile, sc *ScheduleConfig) *Schedule {
	origin := sc.ScheduleOrigin()
	if origin.Type == ScheduleOriginProfile && origin.Name == profile.Name {
		return newSchedule(profile.config, sc, profile)
	}
	panic(fmt.Errorf("invalid use of newScheduleForProfile(%s, %s)", profile.Name, origin))
}

func newSchedule(config *Config, sc *ScheduleConfig, profile *Profile) *Schedule {
	// schedule
	s := new(Schedule)
	if sc != nil {
		s.ScheduleConfig = *sc
	}

	// config
	if config != nil {
		s.ConfigFile = config.GetConfigFile()
	}

	// profile
	var env *util.Environment
	if profile != nil {
		env = profile.GetEnvironment(true)
	}

	// init
	s.init(env)
	return s
}

var uriPrefixRegex = regexp.MustCompile("^(?i)[a-z]{2,}:")

func (s *Schedule) init(env *util.Environment) {
	// fix paths
	rootPath := filepath.Dir(s.ConfigFile)
	s.SystemdDropInFiles = fixPaths(s.SystemdDropInFiles, expandEnv, expandUserHome, absolutePrefix(rootPath))
	if uriPrefixRegex.MatchString(s.Log) {
		s.Log = fixPath(s.Log, expandEnv, expandUserHome)
	} else {
		s.Log = fixPath(s.Log, expandEnv, expandUserHome, absolutePrefix(rootPath))
	}

	// temporary log file
	if s.Log != "" {
		if tempDir, err := util.TempDir(); err == nil && strings.HasPrefix(s.Log, filepath.ToSlash(tempDir)) {
			s.Log = path.Join(constants.TemporaryDirMarker, s.Log[len(tempDir):])
		}
	}

	// capture schedule environment
	if len(s.EnvCapture) > 0 {
		if env == nil {
			env = util.NewDefaultEnvironment(os.Environ()...)
		}

		for index, key := range env.Names() {
			matched := slices.ContainsFunc(s.EnvCapture, func(pattern string) bool {
				matched, err := filepath.Match(pattern, key)
				if err != nil && index == 0 {
					clog.Tracef("env not matched with invalid glob expression '%s': %s", pattern, err.Error())
				}
				return matched
			})
			if !matched {
				env.Remove(key)
			}
		}

		env.Remove(constants.EnvScheduleId)
		s.Environment = env.Values()
	}

	// add the ID of the schedule so that shell hooks can know in which schedule they're in
	s.Environment = append(s.Environment, fmt.Sprintf("%s=%s", constants.EnvScheduleId, s.GetId()))
	sort.Strings(s.Environment)
}

func (s *Schedule) GetId() string {
	return fmt.Sprintf("%s:%s", s.ConfigFile, s.origin)
}

func (s *Schedule) Compare(other *Schedule) (c int) {
	if s == other {
		c = 0
	} else if s == nil {
		c = -1
	} else if other == nil {
		c = 1
	} else {
		c = s.origin.Compare(other.origin)
	}
	return
}

func CompareSchedules(a, b *Schedule) int { return a.Compare(b) }

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
	if !s.LockWait.HasValue() || s.LockWait.Value() <= 2*time.Second {
		return 0
	}
	return s.LockWait.Value()
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
