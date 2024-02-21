package config

import (
	"time"
)

// ScheduleSection contains the information from the schedule profile in the configuration file (v2+).
type ScheduleSection struct {
	config                  *Config
	name                    string
	Group                   string        `mapstructure:"group" description:"Group name to schedule (from groups section)"`
	Profiles                []string      `mapstructure:"profiles" description:"List of profile name to schedule one after another"`
	Command                 string        `mapstructure:"run" default:"backup" examples:"backup;copy;check;forget;prune" description:"Command to schedule. Default is 'backup' if not specified"`
	Schedule                []string      `mapstructure:"schedule" examples:"hourly;daily;weekly;monthly;10:00,14:00,18:00,22:00;Wed,Fri 17:48;*-*-15 02:45;Mon..Fri 00:30" description:"Set the times at which the scheduled command is run (times are specified in systemd timer format)"`
	Permission              string        `mapstructure:"permission" default:"auto" enum:"auto;system;user;user_logged_on" description:"Specify whether the schedule runs with system or user privileges - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	Log                     string        `mapstructure:"log" examples:"/resticprofile.log;tcp://localhost:514" description:"Redirect the output into a log file or to syslog when running on schedule"`
	Priority                string        `mapstructure:"priority" default:"background" enum:"background;standard" description:"Set the priority at which the schedule is run"`
	LockMode                string        `mapstructure:"lock-mode" default:"default" enum:"default;fail;ignore" description:"Specify how locks are used when running on schedule - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	LockWait                time.Duration `mapstructure:"lock-wait" examples:"150s;15m;30m;45m;1h;2h30m" description:"Set the maximum time to wait for acquiring locks when running on schedule"`
	EnvCapture              []string      `mapstructure:"capture-environment" show:"noshow" default:"RESTIC_*" description:"Set names (or glob expressions) of environment variables to capture during schedule creation. The captured environment is applied prior to \"profile.env\" when running the schedule. Whether capturing is supported depends on the type of scheduler being used (supported in \"systemd\" and \"launchd\")"`
	IgnoreOnBattery         bool          `mapstructure:"ignore-on-battery" default:"false" description:"Don't schedule the start of this profile when running on battery"`
	IgnoreOnBatteryLessThan int           `mapstructure:"ignore-on-battery-less-than" default:"" description:"Don't schedule the start of this profile when running on battery, and the battery charge left is less than the value"`
}

// NewScheduleSection instantiates a new blank schedule
func NewScheduleSection(c *Config, name string) *ScheduleSection {
	return &ScheduleSection{
		name:   name,
		config: c,
	}
}

func (s *ScheduleSection) GetSchedule() *Schedule {
	// TODO: implement
	return nil
}

func (s *ScheduleSection) Name() string {
	if len(s.name) == 0 && len(s.Profiles) == 1 {
		// configuration v1
		return s.Profiles[0] + "-" + s.Command
	}
	return s.name
}
