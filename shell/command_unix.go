//+build !windows

package shell

import (
	"syscall"
)

func (c *Command) propagateSignal(pid int) {
	select {
	case <-c.sigChan:
		// We resend the signal to the child process
		syscall.Kill(pid, syscall.SIGINT)
		return
	case <-c.done:
		return
	}
}
