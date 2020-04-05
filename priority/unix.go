//+build !windows

package priority

import (
	"fmt"
	"runtime"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/unix"
)

// SetNice sets the unix "nice" value of the current process
func SetNice(priority int) error {
	var err error
	// pid 0 means "self"
	pid := 0

	if priority < -20 || priority > 19 {
		return fmt.Errorf("Unexpected priority value %d", priority)
	}

	if runtime.GOOS == "linux" {
		// Move ourselves to a new process group so that we can use the process
		// group variants of Setpriority etc to affect all of our threads in one go
		err = unix.Setpgid(pid, 0)
		if err != nil {
			return fmt.Errorf("Error setting process group: %v", err)
		}
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
