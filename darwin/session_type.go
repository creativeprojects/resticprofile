//go:build darwin

package darwin

import "github.com/creativeprojects/resticprofile/constants"

type SessionType string

const (
	SessionTypeDefault    SessionType = ""
	SessionTypeGUI        SessionType = "Aqua"
	SessionTypeBackground SessionType = "Background"
	SessionTypeStandardIO SessionType = "StandardIO"
	SessionTypeSystem     SessionType = "System"
)

func NewSessionType(permission string) SessionType {
	switch permission {
	case constants.SchedulePermissionSystem:
		return SessionTypeSystem

	case constants.SchedulePermissionUser:
		return SessionTypeBackground

	case constants.SchedulePermissionUserLoggedOn, constants.SchedulePermissionUserLoggedIn:
		return SessionTypeGUI

	default:
		// this was the only option available before 0.30.0
		return SessionTypeDefault
	}
}
