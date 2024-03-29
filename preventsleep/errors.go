package preventsleep

import "errors"

var (
	ErrNotSupported     = errors.New("preventing system sleep is not supported on this platform")
	ErrNotStarted       = errors.New("caffeinate is not started")
	ErrAlreadyStarted   = errors.New("caffeinate is already started")
	ErrNotRunning       = errors.New("caffeinate is no longer running")
	ErrPermissionDenied = errors.New("permission denied, you must use sudo to perform this operation")
)
