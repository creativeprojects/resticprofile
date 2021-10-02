package schedule

import (
	"runtime"
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

func TestDetectSchedulePermissionOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	fixtures := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"", "system", true},
		{"something", "system", true},
		{"system", "system", true},
		{"user", "user", true},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			perm, safe := detectSchedulePermission(fixture.input)
			assert.Equal(t, fixture.expected, perm)
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
