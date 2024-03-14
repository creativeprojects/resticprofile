//go:build !windows

package priority

import (
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

	clog.Debugf("setting process group priority to %d", priority)
	err = unix.Setpriority(unix.PRIO_PGRP, selfPID, priority)
	if err != nil {
		return fmt.Errorf("cannot set process group priority, restic will run with the default priority: %w", err)
	}

	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	return SetNice(class)
}
