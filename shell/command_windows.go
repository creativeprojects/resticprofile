//go:build windows

package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// In Windows, all hierarchy will receive the signal (which is good because we cannot send it anyway)
// In fact, there's nothing for us to do here but method must block on channels
func (c *Command) propagateSignal(*os.Process) {
	select {
	case <-c.sigChan:
		return
	case <-c.done:
		return
	}
}

func (c *Command) getShellSearchList() []string {
	return []string{windowsShell, windowsPowershell6, windowsPowershell}
}

func (c *Command) composeShellArguments(shell string) []string {
	command := []string{"/C", c.Command}

	if sh := strings.ToLower(filepath.Base(shell)); sh == windowsPowershell || sh == windowsPowershell6 {
		// Powershell requires ".\" prefix for executables in CWD (same as unix shells)
		if filepath.Base(c.Command) == c.Command {
			if s, err := os.Stat(c.Command); err == nil && !s.IsDir() {
				c.Command = `.\` + c.Command
			}
		}
		command = []string{"-Command", c.Command}
	}

	return append(
		command,
		removeQuotes(c.Arguments)...,
	)
}
