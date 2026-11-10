package schedule

import (
	"fmt"

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

// checkAfterLoginPermission returns an error when the job requests an after-login
// trigger with a permission that cannot be tied to an interactive login session.
// after-login only makes sense for a logged-on user session, so it requires the
// "user_logged_on" permission on every scheduler.
func checkAfterLoginPermission(job *Config, permission Permission) error {
	if job.AfterLogin && permission != PermissionUserLoggedOn {
		return fmt.Errorf("after-login requires the %q permission, but the schedule resolves to %q",
			constants.SchedulePermissionUserLoggedOn, permission.String())
	}
	return nil
}

func (p Permission) String() string {
	switch p {

	case PermissionSystem:
		return constants.SchedulePermissionSystem

	case PermissionUserBackground:
		return constants.SchedulePermissionUser

	case PermissionUserLoggedOn:
		return constants.SchedulePermissionUserLoggedOn

	default:
		return constants.SchedulePermissionAuto
	}
}
