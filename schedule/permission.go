package schedule

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

type Permission int

const (
	PermissionAuto Permission = iota
	PermissionSystem
	PermissionUserBackground
	PermissionUserLoggedOn
)

func PermissionFromConfig(permission string) Permission {
	switch permission {
	case constants.SchedulePermissionSystem:
		return PermissionSystem

	case constants.SchedulePermissionUser:
		return PermissionUserBackground

	case constants.SchedulePermissionUserLoggedIn, constants.SchedulePermissionUserLoggedOn:
		return PermissionUserLoggedOn

	default:
		return PermissionAuto
	}
}

func (p Permission) String() string {
	switch p {
	case PermissionAuto:
		return constants.SchedulePermissionAuto

	case PermissionSystem:
		return constants.SchedulePermissionSystem

	case PermissionUserBackground:
		return constants.SchedulePermissionUser

	case PermissionUserLoggedOn:
		return constants.SchedulePermissionUserLoggedOn

	default:
		return ""
	}
}

func (p Permission) Resolve() Permission {
	permission, safe := p.Detect()
	if !safe {
		clog.Warningf("you have not specified the permission for your schedule (\"system\", \"user\" or \"user_logged_on\"): assuming %q", permission.String())
	}
	return permission
}

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// If the permission can only be guessed, this method will also display a warning
func getSchedulePermission(permission string) string {
	permission, safe := detectSchedulePermission(permission)
	if !safe {
		clog.Warningf("you have not specified the permission for your schedule (\"system\", \"user\" or \"user_logged_on\"): assuming %q", permission)
	}
	return permission
}
