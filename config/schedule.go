package config

import "time"

type Schedule struct {
	Group      string        `mapstructure:"group"`
	Profiles   []string      `mapstructure:"profiles"`
	Command    string        `mapstructure:"run"`
	Schedule   []string      `mapstructure:"schedule"`
	Permission string        `mapstructure:"permission"`
	Logfile    string        `mapstructure:"logfile"`
	Syslog     string        `mapstructure:"syslog"`
	Priority   string        `mapstructure:"priority"`
	LockMode   string        `mapstructure:"lock-mode"`
	LockWait   time.Duration `mapstructure:"lock-wait"`
}
