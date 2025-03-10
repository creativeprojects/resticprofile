//go:build darwin

package darwin

import "github.com/creativeprojects/resticprofile/constants"

type SessionType string

const (
	SessionTypeGUI        SessionType = "Aqua"
	SessionTypeBackground SessionType = "Background"
)

func NewSessionType(permission string) SessionType {
	switch permission {
	case constants.SchedulePermissionSystem:
		return SessionTypeBackground

	case constants.SchedulePermissionUser:
		return SessionTypeBackground

	case constants.SchedulePermissionUserLoggedOn, constants.SchedulePermissionUserLoggedIn:
		return SessionTypeGUI

	default:
		// this was the only option available before 0.30.0
		return SessionTypeGUI
	}
}
