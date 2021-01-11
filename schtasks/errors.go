//+build windows

package schtasks

import "errors"

// Common errors
var (
	ErrorNotRegistered = errors.New("task is not registered")
)
