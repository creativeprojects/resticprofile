//+build linux

package priority

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/unix"
)

const IOPrioClassShift = 13
const IOPrioMask = (1 << IOPrioClassShift) - 1

type IOPrioClass int

const (
	IOPrioClassRT IOPrioClass = iota + 1
	IOPrioClassBE
	IOPrioClassIdle
)

type IOPrioWho int

const (
	IOPrioWhoProcess IOPrioWho = iota + 1
	IOPrioWhoPGRP
	IOPrioWhoUser
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

func GetNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, 0)
	if err != nil {
		return 0, err
	}
	return 20 - pri, nil
}

func SetIOPrio(class IOPrioClass, value int) error {
	if class == IOPrioClassIdle {
		// That's the only valid value for Idle
		value = 0
	}

	_, _, err := unix.Syscall(
		unix.SYS_IOPRIO_SET,
		uintptr(IOPrioWhoPGRP),
		uintptr(os.Getpid()),
		uintptr(class)<<IOPrioClassShift|uintptr(value),
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
	return SetIOPrio(IOPrioClass(class), value)
}

func GetIOPrio(who IOPrioWho) (IOPrioClass, int, error) {
	r1, _, err := unix.Syscall(
		unix.SYS_IOPRIO_GET,
		uintptr(who),
		uintptr(os.Getpid()),
		0,
	)
	if err != 0 {
		return 0, 0, err
	}
	class := r1 >> IOPrioClassShift
	value := r1 & IOPrioMask
	return IOPrioClass(class), int(value), nil
}
