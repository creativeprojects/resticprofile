package schedule

import "errors"

// Generic errors
var (
	ErrScheduledJobNotFound   = errors.New("scheduled job not found")
	ErrScheduledJobNotRunning = errors.New("scheduled job is not running")
)
