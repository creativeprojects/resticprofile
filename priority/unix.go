//go:build !windows && !linux
// +build !windows,!linux

package priority

import (
	"errors"
	"fmt"

	"github.com/creativeprojects/clog"
	"golang.org/x/sys/unix"
)

const selfPID = 0

// SetNice sets the unix "nice" value of the current process
func SetNice(priority int) error {
	var err error

	if priority < -20 || priority > 19 {
		return fmt.Errorf("unexpected priority value %d", priority)
	}

	// Move ourselves to a new process group so that we can use the process
	// group variants of Setpriority to affect all of our processes at once
	err = unix.Setpgid(selfPID, 0)
	if err != nil {
		return fmt.Errorf("cannot set process group, restic will run with the default priority: %w", err)
	}

	clog.Debugf("setting group process priority to %d", priority)
	err = unix.Setpriority(unix.PRIO_PGRP, selfPID, priority)
	if err != nil {
		clog.Debugf("setting process priority to %d instead", priority)
		err = unix.Setpriority(unix.PRIO_PROCESS, selfPID, priority)
		if err != nil {
			return fmt.Errorf("cannot set process priority, restic will run with the default priority: %w", err)
		}
	}
	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	return SetNice(class)
}

// GetProcessNice returns the nice value of the current process
func GetProcessNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, selfPID)
	if err != nil {
		return 0, err
	}
	return pri, nil
}

// GetProcessNice returns the nice value of the current process group
func GetGroupNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PGRP, selfPID)
	if err != nil {
		return 0, err
	}
	return pri, nil
}

// SetIONice does nothing in non-linux OS
func SetIONice(class, value int) error {
	return errors.New("IONice is only supported on Linux")
}
