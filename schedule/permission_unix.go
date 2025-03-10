//go:build !windows

package schedule

import (
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
)

// Detect returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func (p Permission) Detect() (Permission, bool) {
	switch p {
	case PermissionSystem, PermissionUserBackground, PermissionUserLoggedOn:
		// well defined
		return p, true

	default:
		// best guess is depending on the user being root or not:
		detected := PermissionUserLoggedOn // sane default EXCEPT for cron
		if os.Geteuid() == 0 {
			detected = PermissionSystem
		}
		// darwin can backup protected files without the need of a system task
		// otherwise guess based on UID is never safe
		return detected, platform.IsDarwin()
	}
}

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func detectSchedulePermission(permission string) (detected string, safe bool) {
	if permission == constants.SchedulePermissionSystem ||
		permission == constants.SchedulePermissionUser {
		// well defined
		return permission, true
	}
	// best guess is depending on the user being root or not:
	if os.Geteuid() == 0 {
		detected = constants.SchedulePermissionSystem
	} else {
		detected = constants.SchedulePermissionUser
	}
	// darwin can backup protected files without the need of a system task
	// otherwise guess based on UID is never safe
	safe = runtime.GOOS == "darwin"

	return
}

// checkPermission returns true if the user is allowed to access the job.
func checkPermission(permission string) bool {
	if permission == constants.SchedulePermissionUser {
		// user mode is always available
		return true
	}
	if os.Geteuid() == 0 {
		// user has sudoed
		return true
	}
	// last case is system (or undefined) + no sudo
	return false
}

func permissionError(action string) error {
	return fmt.Errorf("user is not allowed to %s a system job: please restart resticprofile as root (with sudo)", action)
}
