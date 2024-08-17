//go:build !windows

package shell

import (
	"os"
	"syscall"
)

func (c *Command) propagateSignal(process *os.Process) {
	select {
	case <-c.sigChan:
		// We resend the signal to the child process
		_ = process.Signal(syscall.SIGINT)
		return
	case <-c.done:
		return
	}
}

// getShellSearchList returns a priority sorted list of default shells to pick when none was specified
func (c *Command) getShellSearchList() []string {
	return []string{
		// prefer "bash" if available as it has better signal propagation (sh may fail to forward signals)
		bashShell,
		// fallback to "sh"
		defaultShell,
	}
}
