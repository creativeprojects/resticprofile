//go:build windows

package schtasks

import "errors"

// Common errors
var (
	ErrNotRegistered = errors.New("task is not registered")
)
