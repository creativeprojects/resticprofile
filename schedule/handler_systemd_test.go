//go:build !darwin && !windows

package schedule

import (
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

	handler := NewHandler(SchedulerSystemd{}).(*HandlerSystemd)

	expectedJobs := []Config{}
	for _, testCase := range testCases {
		expectedJobs = append(expectedJobs, testCase.job)
		toRemove := &testCase.job

		err := handler.CreateJob(&testCase.job, testCase.schedules, schedulePermission)
		t.Cleanup(func() {
			_ = handler.RemoveJob(toRemove, schedulePermission)
		})
		require.NoError(t, err)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedJobs, scheduled)
}
