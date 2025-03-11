//go:build !windows && !darwin

package schedule

import (
	"os"
)

// Check returns true if the user is allowed to access the job.
func (p Permission) Check() bool {
	switch p {
	case PermissionUserLoggedOn:
		// user mode is always available
		return true

	default:
		if os.Geteuid() == 0 {
			// user has sudoed
			return true
		}
		// last case is system (or undefined) + no sudo
		return false

	}
}
