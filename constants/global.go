package constants

import (
	"time"

	"github.com/creativeprojects/resticprofile/priority"
)

// Scheduler type
const (
	SchedulerLaunchd = "launchd"
	SchedulerWindows = "taskscheduler"
	SchedulerSystemd = "systemd"
	SchedulerCrond   = "crond"
)

var (
	// PriorityValues is the map between the name and the value
	PriorityValues = map[string]int{
		"idle":       priority.Idle,
		"background": priority.Background,
		"low":        priority.Low,
		"normal":     priority.Normal,
		"high":       priority.High,
		"highest":    priority.Highest,
	}
)

// Limits for restic lock handling (stale locks and retry on lock failure)
const (
	MinResticLockRetryDelay        = 15 * time.Second
	MaxResticLockRetryDelay        = 30 * time.Minute
	MaxResticLockRetryTimeArgument = 10 * time.Minute
	MinResticStaleLockAge          = 30 * time.Minute
)

// Schedule lock mode config options
const (
	ScheduleLockModeOptionFail   = "fail"
	ScheduleLockModeOptionIgnore = "ignore"
)

const (
	ExitCodeSuccess = 0
	ExitCodeError   = 1
	ExitCodeWarning = 3
)
