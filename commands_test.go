package main

import (
	"bytes"
	"errors"
	"fmt"
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

	// Test that errors from getScheduleJobs are passed through
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
		scheduler := schedule.NewScheduler(schedule.NewHandler(schedule.SchedulerDefaultOS{}), jobConfig.Title)
		defer func(s *schedule.Scheduler) { s.Close() }(scheduler) // Capture current ref to scheduler to be able to close it when function returns.

		if jobConfig.SubTitle == "check" {
			assert.False(t, scheduler.NewJob(jobConfig).RemoveOnly())
			declaredCount++
		} else {
			assert.True(t, scheduler.NewJob(jobConfig).RemoveOnly())
		}
	}

	assert.Equal(t, 1, declaredCount)
}

func TestSchedules(t *testing.T) {
	testConfig := `
[default.check]
schedule = "daily"
[default.backup]
[other.backup]
`
	cfg, err := config.Load(bytes.NewBufferString(testConfig), "toml")
	assert.Nil(t, err)

	// Test that non-existent profiles causes an error
	_, _, _, notFoundErr := getScheduleJobs(cfg, commandLineFlags{name: "non-existent"})
	assert.EqualError(t, notFoundErr, "profile 'non-existent' not found")

	// Test that non-existent schedule causes no error at first
	{
		flags := commandLineFlags{name: "other"}
		_, _, schedules, err := getScheduleJobs(cfg, flags)
		assert.Nil(t, err)

		err = requireScheduleJobs(schedules, flags)
		assert.EqualError(t, err, "no schedule found for profile 'other'")
	}

	// Test that only declared job configs are returned
	{
		flags := commandLineFlags{name: "default"}
		_, profile, schedules, err := getScheduleJobs(cfg, flags)
		assert.Nil(t, err)

		err = requireScheduleJobs(schedules, flags)
		assert.Nil(t, err)

		assert.NotNil(t, profile)
		assert.NotEmpty(t, schedules)
		assert.Len(t, schedules, 1)
		assert.Equal(t, "check", schedules[0].SubTitle)
	}
}

func TestContainsString(t *testing.T) {
	list := []string{"3a", "2b", "1c"}

	tests := map[string]bool{
		"":   false,
		"3b": false,
		"3a": true,
		"2b": true,
		"1c": true,
		"0d": false,
		"1":  false,
	}

	for value, contained := range tests {
		assert.Equal(t, contained, containsString(list, value))
	}

	assert.False(t, containsString(nil, ""))
	assert.False(t, containsString([]string{}, ""))
	assert.True(t, containsString([]string{""}, ""))
}

func TestSelectProfiles(t *testing.T) {
	testConfig := `
[global]
[groups]
others = ["2nd", "3rd"]
default = ["2nd", "default"]  # name collision with default profile
[default.check]
schedule = "daily"
[default.backup]
_ = 0
[2nd.backup]
_ = 0
[3rd.backup]
_ = 0
`
	allProfiles := []string{"default", "2nd", "3rd"}

	cfg, err := config.Load(bytes.NewBufferString(testConfig), "toml")
	assert.Nil(t, err)
	for _, p := range allProfiles {
		assert.True(t, cfg.HasProfile(p))
	}

	// Select --all
	assert.ElementsMatch(t, allProfiles, selectProfiles(cfg, commandLineFlags{}, []string{"--all"}))

	// Select profiles of group
	assert.ElementsMatch(t, []string{"2nd", "3rd"}, selectProfiles(cfg, commandLineFlags{name: "others"}, nil))

	// Select profiles by name
	for _, p := range allProfiles {
		assert.ElementsMatch(t, []string{p}, selectProfiles(cfg, commandLineFlags{name: p}, nil))
	}

	// Select non-existing profile or group
	assert.ElementsMatch(t, []string{"non-existing"}, selectProfiles(cfg, commandLineFlags{name: "non-existing"}, nil))
}

func TestFlagsForProfile(t *testing.T) {
	flags := commandLineFlags{name: "_"}
	profileFlags := flagsForProfile(flags, "test")

	assert.NotEqual(t, flags, profileFlags)
	assert.Equal(t, "_", flags.name)
	assert.Equal(t, "test", profileFlags.name)
}

func TestCompleteCall(t *testing.T) {
	flags := commandLineFlags{}

	completer := &Completer{}
	completer.init(nil)
	newline := fmt.Sprintln("")
	expectedFlags := strings.Join(completer.completeFlagSet(""), newline) + newline

	testTable := []struct {
		args     []string
		expected string
	}{
		{args: []string{"--"}, expected: expectedFlags},
		{args: []string{"bash:v1", "--"}, expected: expectedFlags},
		{args: []string{"bash:v10", "--"}, expected: ""},
		{args: []string{"zsh:v1", "--"}, expected: ""},
	}

	for _, test := range testTable {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			buffer := &strings.Builder{}
			assert.Nil(t, completeCommand(buffer, nil, flags, test.args))
			assert.Equal(t, test.expected, buffer.String())
		})
	}
}

func TestGenerateCommand(t *testing.T) {
	flags := commandLineFlags{}
	buffer := &strings.Builder{}

	t.Run("--bash-completion", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, nil, flags, []string{"--bash-completion"}))
		assert.Equal(t, strings.TrimSpace(bashCompletionScript), strings.TrimSpace(buffer.String()))
		assert.Contains(t, bashCompletionScript, "#!/usr/bin/env bash")
	})

	t.Run("--zsh-completion", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, nil, flags, []string{"--zsh-completion"}))
		assert.Equal(t, strings.TrimSpace(zshCompletionScript), strings.TrimSpace(buffer.String()))
		assert.Contains(t, zshCompletionScript, "#!/usr/bin/env zsh")
	})

	t.Run("--random-key", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, nil, flags, []string{"--random-key", "512"}))
		assert.Equal(t, 684, len(strings.TrimSpace(buffer.String())))
	})

	t.Run("invalid-option", func(t *testing.T) {
		buffer.Reset()
		opts := []string{"", "invalid", "--unknown"}
		for _, option := range opts {
			buffer.Reset()
			err := generateCommand(buffer, nil, flags, []string{option})
			assert.EqualError(t, err, fmt.Sprintf("nothing to generate for: %s", option))
			assert.Equal(t, 0, buffer.Len())
		}
	})
}
