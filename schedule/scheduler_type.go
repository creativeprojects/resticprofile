package schedule

import (
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
)

type SchedulerType interface {
	String() string
}

type SchedulerDefaultOS struct{}

func (s SchedulerDefaultOS) String() string {
	return ""
}

type SchedulerWindows struct{}

func (s SchedulerWindows) String() string {
	return constants.SchedulerWindows
}

type SchedulerLaunchd struct{}

func (s SchedulerLaunchd) String() string {
	return constants.SchedulerLaunchd
}

type SchedulerCrond struct{}

func (s SchedulerCrond) String() string {
	return constants.SchedulerCrond
}

type SchedulerSystemd struct {
	UnitTemplate  string
	TimerTemplate string
}

func (s SchedulerSystemd) String() string {
	return constants.SchedulerSystemd
}

func NewSchedulerType(global *config.Global) SchedulerType {
	switch global.Scheduler {
	case constants.SchedulerCrond:
		return &SchedulerCrond{}
	case constants.SchedulerLaunchd:
		return &SchedulerLaunchd{}
	case constants.SchedulerSystemd:
		return &SchedulerSystemd{
			UnitTemplate:  global.SystemdUnitTemplate,
			TimerTemplate: global.SystemdTimerTemplate,
		}
	case constants.SchedulerWindows:
		return &SchedulerWindows{}
	default:
		return &SchedulerDefaultOS{}
	}
}

var (
	_ SchedulerType = &SchedulerDefaultOS{}
	_ SchedulerType = &SchedulerCrond{}
	_ SchedulerType = &SchedulerLaunchd{}
	_ SchedulerType = &SchedulerSystemd{}
	_ SchedulerType = &SchedulerWindows{}
)
