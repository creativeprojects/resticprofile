package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestDetectSchedulePermissionOnWindows(t *testing.T) {
	if !platform.IsWindows() {
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
		{"user_logged_on", "user_logged_on", true},
		{"user_logged_in", "user_logged_on", true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			perm, safe := detectSchedulePermission(fixture.input)
			assert.Equal(t, fixture.expected, perm)
			assert.Equal(t, fixture.safe, safe)
		})
	}
}

func TestDetectPermission(t *testing.T) {
	fixtures := []struct {
		input    string
		expected string
		safe     bool
		active   bool
	}{
		{"", "system", true, platform.IsWindows()},
		{"something", "system", true, platform.IsWindows()},
		{"", "user_logged_on", platform.IsDarwin(), !platform.IsWindows()},
		{"something", "user_logged_on", platform.IsDarwin(), !platform.IsWindows()},
		{"system", "system", true, true},
		{"user", "user", true, true},
		{"user_logged_on", "user_logged_on", true, true},
		{"user_logged_in", "user_logged_on", true, true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		if !fixture.active {
			continue
		}
		t.Run(fixture.input, func(t *testing.T) {
			perm, safe := PermissionFromConfig(fixture.input).Detect()
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
