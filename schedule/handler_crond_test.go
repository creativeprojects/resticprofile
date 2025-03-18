package schedule

import (
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingCrondScheduled(t *testing.T) {
	hourly := calendar.NewEvent(func(e *calendar.Event) {
		e.Minute.MustAddValue(0)
		e.Second.MustAddValue(0)
	})

	testCases := []struct {
		job       Config
		schedules []*calendar.Event
	}{
		{
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "examples/dev.yaml",
				Permission:       "user",
			},
			schedules: []*calendar.Event{
				hourly,
			},
		},
		{
			job: Config{
				ProfileName:      "test.scheduled",
				CommandName:      "backup",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "config file.yaml", "--name", "test.scheduled", "backup"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "config file.yaml",
				Permission:       "user",
			},
			schedules: []*calendar.Event{
				hourly,
			},
		},
	}

	tempFile := filepath.Join(t.TempDir(), "crontab")
	handler := NewHandler(SchedulerCrond{
		CrontabFile: tempFile,
	}).(*HandlerCrond)
	handler.fs = afero.NewMemMapFs()

	expectedJobs := []Config{}
	for _, testCase := range testCases {
		expectedJobs = append(expectedJobs, testCase.job)

		err := handler.CreateJob(&testCase.job, testCase.schedules, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedJobs, scheduled)

	// now delete all schedules
	for _, testCase := range testCases {
		err := handler.RemoveJob(&testCase.job, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err = handler.Scheduled("")
	require.NoError(t, err)
	assert.Empty(t, scheduled)
}

func TestDetectPermissionCrond(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"", "user", false},
		{"something", "user", false},
		{"system", "system", true},
		{"user", "user", true},
		{"user_logged_on", "user_logged_on", true},
		{"user_logged_in", "user_logged_on", true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(SchedulerCrond{}).(*HandlerCrond)
			perm, safe := handler.DetectSchedulePermission(PermissionFromConfig(fixture.input))
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
