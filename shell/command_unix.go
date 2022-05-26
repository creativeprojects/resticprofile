//go:build !windows

package shell

import (
	"os"
	"strings"
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

func (c *Command) propagateGroupSignal(process *os.Process) {
	select {
	case <-c.sigChan:
		// We resend the signal to the child group
		group, _ := os.FindProcess(-process.Pid)
		group.Signal(syscall.SIGINT)
		return
	case <-c.done:
		return
	}
}

func (c *Command) getShellSearchList() []string {
	// prefer bash if available as it has better signal propagation (sh may fail to forward signals)
	return []string{unixBashShell, unixShell}
}

func (c *Command) composeShellArguments(_ string) []string {
	// Flatten all arguments into one string, sh expects one big string
	flatCommand := append([]string{c.Command}, c.Arguments...)

	return []string{
		"-c",
		strings.Join(flatCommand, " "),
	}
}
