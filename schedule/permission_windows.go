//go:build windows

package schedule

import (
	"errors"

	"github.com/creativeprojects/resticprofile/constants"
)

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func detectSchedulePermission(permission string) (detected string, safe bool) {
	if permission == constants.SchedulePermissionUser {
		return constants.SchedulePermissionUser, true
	}
	return constants.SchedulePermissionSystem, true
}

// checkPermission returns true if the user is allowed to access the job.
// This is always true on Windows
func checkPermission(permission string) bool {
	return true
}

// permissionError is not used in Windows
func permissionError(string) error {
	return errors.New("computer says no")
}
