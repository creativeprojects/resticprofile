//+build linux

package priority

import (
	"fmt"
	"os"

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

	// Move ourselves to a new process group so that we can use the process
	// group variants of Setpriority etc to affect all of our threads in one go
	err = unix.Setpgid(pid, 0)
	if err != nil {
		return fmt.Errorf("Error setting process group: %v", err)
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

const ioPrioClassShift = 13

type ioPrioClass int

const (
	ioPrioClassRT ioPrioClass = iota + 1
	ioPrioClassBE
	ioPrioClassIdle
)

const (
	ioPrioWhoProcess = iota + 1
	ioPrioWhoPGRP
	ioPrioWhoUser
)

func ioPrioSet(class ioPrioClass, value int) error {
	if class == ioPrioClassIdle {
		// That's the only valid value for Idle
		value = 0
	}

	_, _, err := unix.Syscall(
		unix.SYS_IOPRIO_SET,
		uintptr(ioPrioWhoPGRP),
		uintptr(os.Getpid()),
		uintptr(class)<<ioPrioClassShift|uintptr(value),
	)
	if err != 0 {
		return err
	}
	return nil
}

func SetIONice(class, value int) error {
	if class < 0 || class > 3 {
		return fmt.Errorf("SetIONice: expected class to be 1, 2 or 3, found %d", class)
	}
	if value < 0 || value > 7 {
		return fmt.Errorf("SetIONice: expected value from 0 to 7, found %d", value)
	}
	return ioPrioSet(ioPrioClass(class), value)
}
