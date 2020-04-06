//+build !windows,!linux

package priority

import (
	"errors"
	"fmt"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/unix"
)

// SetNice sets the unix "nice" value of the current process
func SetNice(priority int) error {
	var err error
	// pid 0 means "self"
	pid := 0

	if priority < -20 || priority > 19 {
		return fmt.Errorf("Unexpected priority value %d", priority)
	}

	clog.Debugf("Setting process priority to %d", priority)
	err = unix.Setpriority(unix.PRIO_PROCESS, pid, priority)
	if err != nil {
		return fmt.Errorf("Error setting process priority: %v", err)
	}
	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	return SetNice(class)
}

// GetNice returns the scheduler priority of the current process
func GetNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, 0)
	if err != nil {
		return 0, err
	}
	return pri, nil
}

// SetIONice does nothing in non-linux OS
func SetIONice(class, value int) error {
	return errors.New("IONice is only supported on Linux")
}
