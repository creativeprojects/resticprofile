//+build !windows

package priority

import (
	"fmt"

	"github.com/creativeprojects/resticprofile/clog"
	"golang.org/x/sys/unix"
)

func setPriority(priority int) error {
	if priority < 0 || priority > 19 {
		return fmt.Errorf("Unexpected priority value %d", priority)
	}
	pid := unix.Getpid()
	clog.Debugf("Setting priority %d to process %d", priority, pid)
	err := unix.Setpriority(unix.PRIO_PROCESS, pid, priority)
	if err != nil {
		return fmt.Errorf("Error setting process priority: %v", err)
	}
	return nil
}
