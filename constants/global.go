package constants

import (
	"github.com/creativeprojects/resticprofile/priority"
	"time"
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
	MinResticLockRetryTime = 15 * time.Second
	MaxResticLockRetryTime = 30 * time.Minute
	MinResticStaleLockAge  = 1 * time.Hour
)

// Schedule lock mode config options
const (
	ScheduleLockModeOptionFail   = "fail"
	ScheduleLockModeOptionIgnore = "ignore"
)
