//+build windows

package main

import (
	"errors"
	"fmt"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/windows"
)

func setPriority(priority int) error {
	if priority < 0 || priority > 19 {
		return fmt.Errorf("Unexpected priority value %d", priority)
	}
	class := windows.NORMAL_PRIORITY_CLASS
	if priority == 19 {
		class = windows.IDLE_PRIORITY_CLASS
	} else if priority == -20 {
		class = windows.HIGH_PRIORITY_CLASS
	} else if priority > 0 {
		class = BELOW_NORMAL_PRIORITY_CLASS
	} else if priority < 0 {
		class = ABOVE_NORMAL_PRIORITY_CLASS
	}
	handle, err := windows.GetCurrentProcess()
	if err != nil {
		return fmt.Errorf("Error getting current process handle: %v", err)
	}
	clog.Debugf("Setting priority class %d to handle %d", priority, handle)
	err = windows.SetPriorityClass(handle, class)
	if err != nil {
		return fmt.Errorf("Error setting priority class: %v", err)
	}
	return errors.New("Setting priority is not yet supported on Windows")
}
