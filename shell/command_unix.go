//+build !windows

package shell

import (
	"os"
	"syscall"
)

func (c *Command) propagateSignal(process *os.Process) {
	select {
	case <-c.sigChan:
		// We resend the signal to the child process
		process.Signal(syscall.SIGINT)
		return
	case <-c.done:
		return
	}
}
