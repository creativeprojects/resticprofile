//go:build !darwin && !windows

package schedule

import (
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingSystemdScheduled(t *testing.T) {
	event := calendar.NewEvent()
	require.NoError(t, event.Parse("2020-01-01"))

	schedulePermission := constants.SchedulePermissionUser

	testCases := []struct {
		job       Config
		schedules []*calendar.Event
	}{
		{
			job: Config{
				ProfileName:      "testscheduled",
				CommandName:      "backup",
				Command:          "/tmp/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Permission:       schedulePermission,
				ConfigFile:       "examples/dev.yaml",
				Schedules:        []string{event.String()},
			},
			schedules: []*calendar.Event{event},
		},
		{
			job: Config{
				ProfileName:      "test.scheduled",
				CommandName:      "backup",
				Command:          "/tmp/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "config file.yaml", "--name", "self", "backup"}),
				WorkingDirectory: "/resticprofile",
				Permission:       schedulePermission,
				ConfigFile:       "config file.yaml",
				Schedules:        []string{event.String()},
			},
			schedules: []*calendar.Event{event},
		},
	}
	userHome, err := os.UserHomeDir()
	require.NoError(t, err)

	handler := NewHandler(SchedulerSystemd{}).(*HandlerSystemd)

	expectedJobs := []Config{}
	for _, testCase := range testCases {
		job := testCase.job
		err := handler.CreateJob(&job, testCase.schedules, PermissionFromConfig(schedulePermission))

		toRemove := &job
		t.Cleanup(func() {
			_ = handler.RemoveJob(toRemove, PermissionFromConfig(schedulePermission))
		})
		require.NoError(t, err)

		job.Environment = []string{"HOME=" + userHome}
		expectedJobs = append(expectedJobs, job)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	testScheduled := make([]Config, 0, len(scheduled))
	for _, s := range scheduled {
		if s.ConfigFile != "config file.yaml" && s.ConfigFile != "examples/dev.yaml" {
			t.Logf("Ignoring config file %s", s.ConfigFile)
			continue
		}
		testScheduled = append(testScheduled, s)
	}

	assert.ElementsMatch(t, expectedJobs, testScheduled)

	// now delete all schedules
	for _, testCase := range testCases {
		err := handler.RemoveJob(&testCase.job, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err = handler.Scheduled("")
	require.NoError(t, err)
	assert.Empty(t, scheduled)
}
