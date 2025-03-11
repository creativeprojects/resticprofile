package schedule

import (
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
