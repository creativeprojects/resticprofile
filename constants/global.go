package constants

import (
	"time"
)

// Scheduler type
const (
	SchedulerLaunchd   = "launchd"
	SchedulerWindows   = "taskscheduler"
	SchedulerSystemd   = "systemd"
	SchedulerCrond     = "crond"
	SchedulerCrontab   = "crontab"
	SchedulerOSDefault = ""
)

// Limits for restic lock handling (stale locks and retry on lock failure)
const (
	MinResticLockRetryDelay        = 15 * time.Second
	MaxResticLockRetryDelay        = 30 * time.Minute
	MaxResticLockRetryTimeArgument = 10 * time.Minute
	MinResticStaleLockAge          = 15 * time.Minute
)

// Schedule lock mode config options
const (
	ScheduleLockModeOptionFail   = "fail"
	ScheduleLockModeOptionIgnore = "ignore"
)

// Exit codes from restic
const (
	ResticExitCodeSuccess            = 0
	ResticExitCodeError              = 1
	ResticExitCodeGoRuntimeError     = 2
	ResticExitCodeWarning            = 3
	ResticExitCodeNoRepository       = 10
	ResticExitCodeFailLockRepository = 11
	ResticExitCodeWrongPassword      = 12
	ResticExitCodeInterrupted        = 130
)
