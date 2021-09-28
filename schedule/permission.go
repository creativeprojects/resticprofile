package schedule

import "github.com/creativeprojects/resticprofile/constants"

// Permission is either system or user
type Permission int

// Permission
const (
	PermissionUser Permission = iota
	PermissionSystem
)

// String returns either "user" or "system"
func (p Permission) String() string {
	if p == PermissionSystem {
		return constants.SchedulePermissionSystem
	}
	return constants.SchedulePermissionUser
}
