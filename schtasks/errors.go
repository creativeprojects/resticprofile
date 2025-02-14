package schtasks

import "errors"

// Common errors
var (
	ErrNotRegistered   = errors.New("task is not registered")
	ErrEmptyTaskName   = errors.New("task name cannot be empty")
	ErrInvalidTaskName = errors.New("invalid task name")
	ErrAccessDenied    = errors.New("access denied")
)
