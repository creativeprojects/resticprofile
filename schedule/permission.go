package schedule

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

// Permission is either system or user
type Permission string

// Permission
const (
	PermissionUser   Permission = "user"
	PermissionSystem Permission = "system"
)

// String returns either "user" or "system"
func (p Permission) String() string {
	if p == PermissionSystem {
		return constants.SchedulePermissionSystem
	}
	return constants.SchedulePermissionUser
}

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// If the permission can only be guessed, this method will also display a warning
func getSchedulePermission(permission string) string {
	permission, safe := detectSchedulePermission(permission)
	if !safe {
		clog.Warningf("you have not specified the permission for your schedule (\"system\" or \"user\"): assuming %q", permission)
	}
	return permission
}
