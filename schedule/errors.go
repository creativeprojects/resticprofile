package schedule

import "errors"

// Generic errors
var (
	ErrServiceNotFound   = errors.New("service not found")
	ErrServiceNotRunning = errors.New("service is not running")
)
