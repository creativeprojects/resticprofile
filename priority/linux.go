//go:build linux

package priority

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/creativeprojects/clog"
	"golang.org/x/sys/unix"
)

const IOPrioClassShift = 13
const IOPrioMask = (1 << IOPrioClassShift) - 1

type IOPrioClass int

// IOPrioClass
const (
	IOPrioNoClass IOPrioClass = iota
	IOPrioClassRT
	IOPrioClassBE
	IOPrioClassIdle
)

type IOPrioWho int

// IOPrioWho
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
		return fmt.Errorf("unexpected priority value %d", priority)
	}

	currentPriority, _ := unix.Getpriority(unix.PRIO_PROCESS, 0)
	if currentPriority == priority {
		return nil
	}

	// Move ourselves to a new process group so that we can use the process
	// group variants of Setpriority etc to affect all of our threads in one go
	err = unix.Setpgid(pid, 0)
	if err != nil {
		return fmt.Errorf("cannot set process group priority, restic will run with the default priority: %w", err)
	}

	clog.Debugf("setting process priority to %d", priority)
	err = unix.Setpriority(unix.PRIO_PROCESS, pid, priority)
	if err != nil {
		return fmt.Errorf("cannot set process priority, restic will run with the default priority: %w", err)
	}
	return nil
}

// SetClass sets the priority class of the current process
func SetClass(class int) error {
	return SetNice(class)
}

// GetNice returns the current nice value
func GetNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, 0)
	if err != nil {
		return 0, err
	}
	return 20 - pri, nil
}

// SetIONice sets the io_prio class and value
func SetIONice(class, value int) error {
	if class < 1 || class > 3 {
		return fmt.Errorf("SetIONice: expected class to be 1, 2 or 3, found %d", class)
	}
	if value < 0 || value > 7 {
		return fmt.Errorf("SetIONice: expected value from 0 to 7, found %d", value)
	}
	clog.Debugf("setting IO priority class to %d, level %d", class, value)
	// Try group of processes first
	err := setIOPrio(IOPrioWhoPGRP, IOPrioClass(class), value)
	if err != nil {
		// Try process only
		return setIOPrio(IOPrioWhoProcess, IOPrioClass(class), value)
	}
	return nil
}

// GetIONice returns the io_prio class and value
func GetIONice() (IOPrioClass, int, error) {
	class, value, err := getIOPrio(IOPrioWhoPGRP)
	if err != nil {
		class, value, err = getIOPrio(IOPrioWhoProcess)
	}
	return class, value, err
}

func getIOPrio(who IOPrioWho) (IOPrioClass, int, error) {

	r1, _, errno := unix.Syscall(
		unix.SYS_IOPRIO_GET,
		uintptr(who),
		uintptr(os.Getpid()),
		0,
	)
	if errno != 0 {
		return 0, 0, errnoToError(errno)
	}
	class := r1 >> IOPrioClassShift
	value := r1 & IOPrioMask
	return IOPrioClass(class), int(value), nil
}

func setIOPrio(who IOPrioWho, class IOPrioClass, value int) error {

	_, _, errno := unix.Syscall(
		unix.SYS_IOPRIO_SET,
		uintptr(who),
		uintptr(os.Getpid()),
		uintptr(class)<<IOPrioClassShift|uintptr(value),
	)
	if errno != 0 {
		return errnoToError(errno)
	}
	return nil
}

func errnoToError(errno syscall.Errno) error {
	message := "unspecified error"

	switch errno {
	case unix.EINVAL:
		message = "invalid class or value"
	case unix.EPERM:
		message = "permission denied"
	case unix.ESRCH:
		message = "process, group of processes, or user processes not found"
	}
	return errors.New(message)
}
