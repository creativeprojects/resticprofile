package schedule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type SchedulerConfig interface {
	// Type of scheduler config ("windows", "launchd", "crond", "systemd" or "" for OS default)
	Type() string
	Convert(typeName string) SchedulerConfig
}

type SchedulerDefaultOS struct {
	defaults []SchedulerConfig
}

func (s SchedulerDefaultOS) Type() string { return constants.SchedulerOSDefault }
func (s SchedulerDefaultOS) Convert(typeName string) SchedulerConfig {
	for _, c := range s.defaults {
		if c.Type() == typeName {
			return c
		}
	}
	return s
}

type SchedulerWindows struct{}

func (s SchedulerWindows) Type() string                     { return constants.SchedulerWindows }
func (s SchedulerWindows) Convert(_ string) SchedulerConfig { return s }

type SchedulerLaunchd struct{}

func (s SchedulerLaunchd) Type() string                     { return constants.SchedulerLaunchd }
func (s SchedulerLaunchd) Convert(_ string) SchedulerConfig { return s }

// SchedulerCrond configures crond compatible schedulers, either needs CrontabBinary or CrontabFile
type SchedulerCrond struct {
	CrontabFile   string
	CrontabBinary string
	Username      string
}

func (s SchedulerCrond) Type() string                     { return constants.SchedulerCrond }
func (s SchedulerCrond) Convert(_ string) SchedulerConfig { return s }

type SchedulerSystemd struct {
	UnitTemplate  string
	TimerTemplate string
	Nice          int
	IONiceClass   int
	IONiceLevel   int
}

func (s SchedulerSystemd) Type() string                     { return constants.SchedulerSystemd }
func (s SchedulerSystemd) Convert(_ string) SchedulerConfig { return s }

func NewSchedulerConfig(global *config.Global) SchedulerConfig {
	// scheduler: resource
	scheduler, resource, _ := strings.Cut(global.Scheduler, ":")
	scheduler = strings.TrimSpace(scheduler)
	resource = strings.TrimSpace(resource)

	switch scheduler {
	case constants.SchedulerCrond:
		if len(resource) > 0 {
			return SchedulerCrond{CrontabBinary: resource}
		} else {
			return SchedulerCrond{}
		}

	case constants.SchedulerCrontab:
		if len(resource) > 0 {
			if user, location, found := strings.Cut(resource, ":"); found {
				user = strings.TrimSpace(user)
				if !regexp.MustCompile(`^[A-Za-z]$`).MatchString(user) { // Checking the username is not a single letter?
					if user == "" {
						user = "-"
					}
					return SchedulerCrond{CrontabFile: strings.TrimSpace(location), Username: user}
				}
			}
			return SchedulerCrond{CrontabFile: resource}
		} else {
			panic(fmt.Errorf("invalid schedule %q, no crontab file was specified, expecting \"%s: filename\"", scheduler, scheduler))
		}

	case constants.SchedulerLaunchd:
		return SchedulerLaunchd{}

	case constants.SchedulerSystemd:
		return getSchedulerSystemdDefaultConfig(global)

	case constants.SchedulerWindows:
		return SchedulerWindows{}

	default:
		return SchedulerDefaultOS{
			defaults: []SchedulerConfig{
				getSchedulerSystemdDefaultConfig(global),
			},
		}
	}
}

func getSchedulerSystemdDefaultConfig(global *config.Global) SchedulerSystemd {
	scheduler := SchedulerSystemd{
		UnitTemplate:  global.SystemdUnitTemplate,
		TimerTemplate: global.SystemdTimerTemplate,
		Nice:          global.Nice,
	}
	if global.IONice {
		scheduler.IONiceClass = global.IONiceClass
		scheduler.IONiceLevel = global.IONiceLevel
	}
	return scheduler
}

var (
	_ SchedulerConfig = SchedulerDefaultOS{}
	_ SchedulerConfig = SchedulerCrond{}
	_ SchedulerConfig = SchedulerLaunchd{}
	_ SchedulerConfig = SchedulerSystemd{}
	_ SchedulerConfig = SchedulerWindows{}
)
