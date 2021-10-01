package schedule

import (
	"runtime"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type SchedulerConfig interface {
	// Type of scheduler config ("windows", "launchd", "crond", "systemd" or "" for OS default)
	Type() string
}

type SchedulerDefaultOS struct{}

func (s SchedulerDefaultOS) Type() string {
	return ""
}

type SchedulerWindows struct{}

func (s SchedulerWindows) Type() string {
	return constants.SchedulerWindows
}

type SchedulerLaunchd struct{}

func (s SchedulerLaunchd) Type() string {
	return constants.SchedulerLaunchd
}

type SchedulerCrond struct{}

func (s SchedulerCrond) Type() string {
	return constants.SchedulerCrond
}

type SchedulerSystemd struct {
	UnitTemplate  string
	TimerTemplate string
}

func (s SchedulerSystemd) Type() string {
	return constants.SchedulerSystemd
}

func NewSchedulerConfig(global *config.Global) SchedulerConfig {
	switch global.Scheduler {
	case constants.SchedulerCrond:
		return SchedulerCrond{}
	case constants.SchedulerLaunchd:
		return SchedulerLaunchd{}
	case constants.SchedulerSystemd:
		return SchedulerSystemd{
			UnitTemplate:  global.SystemdUnitTemplate,
			TimerTemplate: global.SystemdTimerTemplate,
		}
	case constants.SchedulerWindows:
		return SchedulerWindows{}
	default:
		if runtime.GOOS != "darwin" && runtime.GOOS != "windows" {
			return SchedulerSystemd{
				UnitTemplate:  global.SystemdUnitTemplate,
				TimerTemplate: global.SystemdTimerTemplate,
			}
		}
		return SchedulerDefaultOS{}
	}
}

var (
	_ SchedulerConfig = SchedulerDefaultOS{}
	_ SchedulerConfig = SchedulerCrond{}
	_ SchedulerConfig = SchedulerLaunchd{}
	_ SchedulerConfig = SchedulerSystemd{}
	_ SchedulerConfig = SchedulerWindows{}
)
