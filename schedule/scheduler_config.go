package schedule

import (
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
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

type SchedulerLaunchd struct {
	Fs afero.Fs
}

func (s SchedulerLaunchd) Type() string {
	return constants.SchedulerLaunchd
}

type SchedulerCrond struct {
	Fs afero.Fs
}

func (s SchedulerCrond) Type() string {
	return constants.SchedulerCrond
}

type SchedulerSystemd struct {
	Fs            afero.Fs
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
		if !platform.IsDarwin() && !platform.IsWindows() {
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
