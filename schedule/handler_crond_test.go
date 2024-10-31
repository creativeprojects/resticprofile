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
				Schedules:        []string{"*-*-* *:*:*"}, // no parsing of crontab schedules
				ConfigFile:       "examples/dev.yaml",
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
				Schedules:        []string{"*-*-* *:*:*"}, // no parsing of crontab schedules
				ConfigFile:       "config file.yaml",
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

		err := handler.CreateJob(&testCase.job, testCase.schedules, testCase.job.Permission)
		require.NoError(t, err)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedJobs, scheduled)
}

func TestSplitArguments(t *testing.T) {
	testCases := []struct {
		commandLine  string
		expectedArgs []string
	}{
		{
			commandLine:  `cmd arg1 arg2`,
			expectedArgs: []string{"cmd", "arg1", "arg2"},
		},
		{
			commandLine:  `cmd "arg with spaces" arg3`,
			expectedArgs: []string{"cmd", "arg with spaces", "arg3"},
		},
		{
			commandLine:  `cmd "arg with spaces" "another arg"`,
			expectedArgs: []string{"cmd", "arg with spaces", "another arg"},
		},
		{
			commandLine:  `cmd "arg with spaces"`,
			expectedArgs: []string{"cmd", "arg with spaces"},
		},
		{
			commandLine:  `cmd`,
			expectedArgs: []string{"cmd"},
		},
		{
			commandLine:  `"cmd file"`,
			expectedArgs: []string{"cmd file"},
		},
		{
			commandLine:  `"cmd file" arg`,
			expectedArgs: []string{"cmd file", "arg"},
		},
	}

	for _, testCase := range testCases {
		args := splitArguments(testCase.commandLine)
		assert.Equal(t, testCase.expectedArgs, args)
	}
}
