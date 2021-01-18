//+build !windows

package schedule

import (
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
//
// This method is for Unixes only
func (j *Job) getSchedulePermission() string {
	const message = "you have not specified the permission for your schedule (system or user): assuming "
	if j.config.Permission() == constants.SchedulePermissionSystem ||
		j.config.Permission() == constants.SchedulePermissionUser {
		// well defined
		return j.config.Permission()
	}
	// best guess is depending on the user being root or not:
	if os.Geteuid() == 0 {
		if runtime.GOOS != "darwin" {
			// darwin can backup protected files without the need of a system task; no need to bother the user then
			clog.Warning(message, "system")
		}
		return constants.SchedulePermissionSystem
	}
	if runtime.GOOS != "darwin" {
		// darwin can backup protected files without the need of a system task; no need to bother the user then
		clog.Warning(message, "user")
	}
	return constants.SchedulePermissionUser
}

// checkPermission returns true if the user is allowed.
//
// This method is for Unixes only
func (j *Job) checkPermission(permission string) bool {
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
