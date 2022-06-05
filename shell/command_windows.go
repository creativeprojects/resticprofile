//go:build windows

package shell

import (
	"os"
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

// getShellSearchList returns a priority sorted list of default shells to pick when none was specified
func (c *Command) getShellSearchList() []string {
	return []string{
		// prefer "cmd.exe" over "powershell.exe"
		windowsShell,
		// Should never come to here as "cmd.exe" probably always exists
		powershell,
	}
}
