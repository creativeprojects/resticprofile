package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestPermissionUserString(t *testing.T) {
	permission := PermissionUser
	assert.Equal(t, constants.SchedulePermissionUser, permission.String())
}

func TestPermissionSystemString(t *testing.T) {
	permission := PermissionSystem
	assert.Equal(t, constants.SchedulePermissionSystem, permission.String())
}
