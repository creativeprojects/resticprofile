//go:build windows

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerCrond(t *testing.T) {
	handler := NewHandler(SchedulerCrond{})
	assert.IsType(t, &HandlerCrond{}, handler)
}

func TestHandlerDefaultOS(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.IsType(t, &HandlerWindows{}, handler)
}

func TestDetectPermissionTaskScheduler(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			handler := NewHandler(SchedulerWindows{}).(*HandlerWindows)
			perm, safe := handler.DetectSchedulePermission(PermissionFromConfig(fixture.input))
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
