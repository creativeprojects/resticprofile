package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Change default commands for testing ones
	ownCommands = []ownCommand{
		{
			name:              "first",
			description:       "first first",
			action:            firstCommand,
			needConfiguration: false,
		},
		{
			name:              "second",
			description:       "second second",
			action:            secondCommand,
			needConfiguration: true,
		},
		{
			name:              "third",
			description:       "third third",
			action:            thirdCommand,
			needConfiguration: false,
			hide:              true,
		},
	}
}

func firstCommand(_ io.Writer, _ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("first")
}

func secondCommand(_ io.Writer, _ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("second")
}

func thirdCommand(_ io.Writer, _ *config.Config, _ commandLineFlags, _ []string) error {
	return errors.New("third")
}

func TestDisplayOwnCommands(t *testing.T) {
	buffer := &strings.Builder{}
	displayOwnCommands(buffer)
	assert.Equal(t, "   first    first first\n   second   second second\n", buffer.String())
}

func TestIsOwnCommand(t *testing.T) {
	assert.True(t, isOwnCommand("first", false))
	assert.True(t, isOwnCommand("second", true))
	assert.True(t, isOwnCommand("third", false))
	assert.False(t, isOwnCommand("another one", true))
}

func TestRunOwnCommand(t *testing.T) {
	assert.EqualError(t, runOwnCommand(nil, "first", commandLineFlags{}, nil), "first")
	assert.EqualError(t, runOwnCommand(nil, "second", commandLineFlags{}, nil), "second")
	assert.EqualError(t, runOwnCommand(nil, "third", commandLineFlags{}, nil), "third")
	assert.EqualError(t, runOwnCommand(nil, "another one", commandLineFlags{}, nil), "command not found: another one")
}

func TestPanicCommand(t *testing.T) {
	assert.Panics(t, func() {
		_ = panicCommand(nil, nil, commandLineFlags{}, nil)
	})
}

func TestRandomKeyOfInvalidSize(t *testing.T) {
	assert.Error(t, randomKey(os.Stdout, nil, commandLineFlags{resticArgs: []string{"restic", "size"}}, nil))
}

func TestRandomKeyOfZeroSize(t *testing.T) {
	assert.Error(t, randomKey(os.Stdout, nil, commandLineFlags{resticArgs: []string{"restic", "0"}}, nil))
}

func TestRandomKey(t *testing.T) {
	// doesn't look like much, but it's testing the random generator is not throwing an error
	assert.NoError(t, randomKey(os.Stdout, nil, commandLineFlags{}, nil))
}

func TestRemovableSchedules(t *testing.T) {
	testConfig := `
[default.check]
schedule = "daily"
[default.backup]
`
	parsedConfig, err := config.Load(bytes.NewBufferString(testConfig), "toml")
	assert.Nil(t, err)

	// Test that errors from getScheduleJobs are passed thru
	_, _, _, notFoundErr := getRemovableScheduleJobs(parsedConfig, commandLineFlags{name: "non-existent"})
	assert.EqualError(t, notFoundErr, "profile 'non-existent' not found")

	// Test that declared and declarable job configs are returned
	_, profile, schedules, err := getRemovableScheduleJobs(parsedConfig, commandLineFlags{name: "default"})
	assert.Nil(t, err)
	assert.NotNil(t, profile)
	assert.NotEmpty(t, schedules)
	assert.Len(t, schedules, len(profile.SchedulableCommands()))

	declaredCount := 0

	for _, jobConfig := range schedules {
		scheduler := schedule.NewScheduler("", jobConfig.Title())
		defer func(s schedule.Scheduler) { s.Close() }(scheduler) // Capture current ref to scheduler to be able to close it when function returns.

		if jobConfig.SubTitle() == "check" {
			assert.False(t, scheduler.NewJob(jobConfig).RemoveOnly())
			declaredCount++
		} else {
			assert.True(t, scheduler.NewJob(jobConfig).RemoveOnly())
		}
	}

	assert.Equal(t, 1, declaredCount)
}
