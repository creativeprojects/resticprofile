//go:build windows

package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Support for Windows removed as it was broken
// func TestHandlerCrond(t *testing.T) {
// 	handler := NewHandler(SchedulerCrond{})
// 	assert.IsType(t, &HandlerCrond{}, handler)
// }

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

func TestHideWindowOption(t *testing.T) {
	job := Config{
		ProfileName:      "TestHideWindowOption",
		CommandName:      "backup",
		Command:          "echo",
		Arguments:        NewCommandArguments([]string{"hello", "there"}),
		WorkingDirectory: "C:\\",
		JobDescription:   "TestHideWindowOption",
		HideWindow:       true,
	}

	handler := NewHandler(SchedulerWindows{}).(*HandlerWindows)

	event := calendar.NewEvent()
	err := event.Parse("2020-01-02 03:04") // will never get triggered
	require.NoError(t, err)

	err = handler.CreateJob(&job, []*calendar.Event{event}, PermissionUserLoggedOn)
	assert.NoError(t, err)
	defer func() {
		_ = handler.RemoveJob(&job, PermissionUserLoggedOn)
	}()

	scheduledJobs, err := handler.Scheduled(job.ProfileName)
	assert.NoError(t, err)
	assert.Equal(t, len(scheduledJobs), 1)

	assert.Equal(t, scheduledJobs[0].Command, "conhost.exe")
	assert.Equal(t, scheduledJobs[0].Arguments.String(), "--headless echo hello there")
}

func TestStartWhenAvailableOption(t *testing.T) {
	job := Config{
		ProfileName:        "TestStartWhenAvailableOption",
		CommandName:        "backup",
		Command:            "echo",
		Arguments:          NewCommandArguments([]string{"hello", "there"}),
		WorkingDirectory:   "C:\\",
		JobDescription:     "TestStartWhenAvailableOption",
		StartWhenAvailable: true,
	}

	handler := NewHandler(SchedulerWindows{}).(*HandlerWindows)

	event := calendar.NewEvent()
	err := event.Parse("2020-01-02 03:04") // will never get triggered
	require.NoError(t, err)

	err = handler.CreateJob(&job, []*calendar.Event{event}, PermissionUserLoggedOn)
	assert.NoError(t, err)
	defer func() {
		_ = handler.RemoveJob(&job, PermissionUserLoggedOn)
	}()

	scheduledJobs, err := handler.Scheduled(job.ProfileName)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(scheduledJobs))
}
