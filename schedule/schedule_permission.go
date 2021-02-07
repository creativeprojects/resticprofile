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
	permission, unsafe := j.detectSchedulePermission()
	if unsafe {
		clog.Warningf("you have not specified the permission for your schedule (system or user): assuming %s", permission)
	}
	return permission
}

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// unsafe specifies whether a guess may lead to a too broad or too narrow file access permission.
//
// This method is for Unixes only
func (j *Job) detectSchedulePermission() (permission string, unsafe bool) {
	if j.config.Permission() == constants.SchedulePermissionSystem ||
		j.config.Permission() == constants.SchedulePermissionUser {
		// well defined
		permission = j.config.Permission()
		unsafe = false

	} else {
		// best guess is depending on the user being root or not:
		if os.Geteuid() == 0 {
			permission = constants.SchedulePermissionSystem
		} else {
			permission = constants.SchedulePermissionUser
		}
		// darwin can backup protected files without the need of a system task; Guess based on UID is never unsafe
		unsafe = runtime.GOOS != "darwin"
	}

	return
}

// checkPermission returns true if the user is allowed to access the job.
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
