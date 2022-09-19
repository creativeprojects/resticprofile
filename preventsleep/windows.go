//go:build windows

package preventsleep

import (
	"syscall"
)

// Execution States
const (
	EsSystemRequired = 0x00000001
	EsContinuous     = 0x80000000
)

const (
	operationCompleted = "The operation completed successfully."
)

type Caffeinate struct {
	setThreadExecStateProc *syscall.LazyProc
	running                bool
}

func New() *Caffeinate {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setThreadExecStateProc := kernel32.NewProc("SetThreadExecutionState")
	return &Caffeinate{
		setThreadExecStateProc: setThreadExecStateProc,
	}
}

func (c *Caffeinate) Start() error {
	if c.IsRunning() {
		return ErrAlreadyStarted
	}
	var err error
	_, _, err = c.setThreadExecStateProc.Call(uintptr(EsSystemRequired | EsContinuous))
	if err != nil && err.Error() != operationCompleted {
		return err
	}
	c.running = true
	return nil
}

func (c *Caffeinate) Stop() error {
	if !c.IsRunning() {
		return ErrNotStarted
	}
	var err error
	_, _, err = c.setThreadExecStateProc.Call(uintptr(EsContinuous))
	if err != nil && err.Error() != operationCompleted {
		return err
	}
	c.running = false
	return nil
}

func (c *Caffeinate) IsRunning() bool {
	return c.running
}
