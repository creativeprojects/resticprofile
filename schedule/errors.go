package schedule

import "errors"

// Generic errors
var (
	ErrorServiceNotFound   = errors.New("service not found")
	ErrorServiceNotRunning = errors.New("service is not running")
)
