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
	class := IOPrioClass(r1 >> IOPrioClassShift)
	value := int(r1 & IOPrioMask)
	return class, value, nil
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
