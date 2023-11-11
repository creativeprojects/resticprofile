package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/stretchr/testify/assert"
)

func TestPanicCommand(t *testing.T) {
	assert.Panics(t, func() {
		_ = panicCommand(nil, commandContext{})
	})
}

func TestRandomKeyOfInvalidSize(t *testing.T) {
	assert.Error(t, randomKey(os.Stdout, commandContext{
		Context: Context{
			flags: commandLineFlags{resticArgs: []string{"restic", "size"}},
		},
	}))
}

func TestRandomKeyOfZeroSize(t *testing.T) {
	assert.Error(t, randomKey(os.Stdout, commandContext{
		Context: Context{
			flags: commandLineFlags{resticArgs: []string{"restic", "0"}},
		},
	}))
}

func TestRandomKey(t *testing.T) {
	// doesn't look like much, but it's testing the random generator is not throwing an error
	assert.NoError(t, randomKey(os.Stdout, commandContext{}))
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
		scheduler := schedule.NewScheduler(schedule.NewHandler(schedule.SchedulerDefaultOS{}), jobConfig.Profiles[0])
		defer func(s *schedule.Scheduler) { s.Close() }(scheduler) // Capture current ref to scheduler to be able to close it when function returns.

		if jobConfig.CommandName == "check" {
			assert.False(t, scheduler.NewJob(scheduleToConfig(jobConfig)).RemoveOnly())
			declaredCount++
		} else {
			assert.True(t, scheduler.NewJob(scheduleToConfig(jobConfig)).RemoveOnly())
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
		assert.Equal(t, "check", schedules[0].CommandName)
	}
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
	completer := NewCompleter(ownCommands.All(), DefaultFlagsLoader)
	completer.init(nil)
	newline := fmt.Sprintln("")
	expectedFlags := strings.Join(completer.completeFlagSet(""), newline) + newline

	visibleCommands := collect.Not(func(c ownCommand) bool { return c.hideInCompletion || c.hide })
	commandName := func(c ownCommand) string { return c.name }
	commandNames := collect.From(collect.All(ownCommands.All(), visibleCommands), commandName)
	sort.Strings(commandNames)
	expectedCommands := strings.Join(commandNames, newline) + newline +
		RequestResticCompletion + newline

	testTable := []struct {
		args     []string
		expected string
	}{
		{args: []string{"--"}, expected: expectedFlags},
		{args: []string{"--config=does-not-exist", ""}, expected: expectedCommands},
		{args: []string{"bash:v1", "--"}, expected: expectedFlags},
		{args: []string{"bash:v10", "--"}, expected: ""},
		{args: []string{"zsh:v1", "--"}, expected: ""},
	}

	for _, test := range testTable {
		t.Run(strings.Join(test.args, " "), func(t *testing.T) {
			buffer := &strings.Builder{}
			assert.Nil(t, completeCommand(buffer, commandContext{
				ownCommands: ownCommands,
				Context:     Context{request: Request{arguments: test.args}},
			}))
			assert.Equal(t, test.expected, buffer.String())
		})
	}
}

func TestGenerateCommand(t *testing.T) {
	buffer := &strings.Builder{}

	contextWithArguments := func(args []string) commandContext {
		t.Helper()
		return commandContext{Context: Context{request: Request{arguments: args}}} //nolint:exhaustivestruct
	}

	t.Run("--bash-completion", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, contextWithArguments([]string{"--bash-completion"})))
		assert.Equal(t, strings.TrimSpace(bashCompletionScript), strings.TrimSpace(buffer.String()))
		assert.Contains(t, bashCompletionScript, "#!/usr/bin/env bash")
	})

	t.Run("--zsh-completion", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, contextWithArguments([]string{"--zsh-completion"})))
		assert.Equal(t, strings.TrimSpace(zshCompletionScript), strings.TrimSpace(buffer.String()))
		assert.Contains(t, zshCompletionScript, "#!/usr/bin/env zsh")
	})

	t.Run("--config-reference", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--config-reference"})))
		ref := buffer.String()
		assert.Contains(t, ref, "| **ionice-class** |")
		assert.Contains(t, ref, "| **check-after** |")
		assert.Contains(t, ref, "| **continue-on-error** |")
	})

	t.Run("--json-schema", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema"})))
		ref := buffer.String()
		assert.Contains(t, ref, "\"profiles\":")
		assert.Contains(t, ref, "/jsonschema/config-2.json")
	})

	t.Run("--json-schema v1", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "v1"})))
		ref := buffer.String()
		assert.Contains(t, ref, "/jsonschema/config-1.json")
	})

	t.Run("--json-schema --version 0.13 v1", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "--version", "0.13", "v1"})))
		ref := buffer.String()
		assert.Contains(t, ref, "/jsonschema/config-1-restic-0-13.json")
	})

	t.Run("--random-key", func(t *testing.T) {
		buffer.Reset()
		assert.Nil(t, generateCommand(buffer, contextWithArguments([]string{"--random-key", "512"})))
		assert.Equal(t, 684, len(strings.TrimSpace(buffer.String())))
	})

	t.Run("invalid-option", func(t *testing.T) {
		buffer.Reset()
		opts := []string{"", "invalid", "--unknown"}
		for _, option := range opts {
			buffer.Reset()
			err := generateCommand(buffer, contextWithArguments([]string{option}))
			assert.EqualError(t, err, fmt.Sprintf("nothing to generate for: %s", option))
			assert.Equal(t, 0, buffer.Len())
		}
	})
}

func TestCommandFilter(t *testing.T) {
	p, g := new(config.Profile), new(config.Global)

	reset := func() { p.AllowedCommands, p.DeniedCommands, g.BaseProfiles = nil, nil, nil }

	expect := func(t *testing.T, expected bool, command string) {
		assert.Equal(t, expected, commandFilter(p, g)(command), "command %s", command)
	}

	t.Run("nil-tolerant", func(t *testing.T) {
		assert.True(t, commandFilter(nil, nil)("backup"))
	})

	t.Run("default-base", func(t *testing.T) {
		reset()
		p.Name = "default"
		assert.True(t, commandFilter(p, nil)("backup"))
		expect(t, true, "backup")
		p.Name = "__default"
		assert.False(t, commandFilter(p, nil)("backup"))
		expect(t, false, "backup")
	})

	t.Run("configured-base", func(t *testing.T) {
		reset()
		p.Name = "default"
		g.BaseProfiles = []string{"*"}
		expect(t, false, "backup")
		g.BaseProfiles = []string{"default"}
		expect(t, false, "backup")
		p.Name = "other"
		expect(t, true, "backup")
	})

	t.Run("default-all-allowed", func(t *testing.T) {
		reset()
		for run := 0; run < 3; run++ {
			expect(t, true, "backup")
			expect(t, true, "restore")
			expect(t, true, "another")

			switch run {
			case 0:
				p.AllowedCommands, p.DeniedCommands = []string{}, []string{}
			case 1:
				p.AllowedCommands, p.DeniedCommands = []string{""}, []string{""}
			}
		}
	})

	t.Run("allowed", func(t *testing.T) {
		reset()
		p.AllowedCommands = []string{"*"}
		expect(t, true, "backup")
		expect(t, true, "restore")
		p.AllowedCommands = []string{"backup"}
		expect(t, true, "backup")
		expect(t, false, "restore")
	})

	t.Run("denied", func(t *testing.T) {
		reset()
		p.DeniedCommands = []string{"*"}
		expect(t, false, "backup")
		expect(t, false, "restore")
		p.DeniedCommands = []string{"backup"}
		expect(t, false, "backup")
		expect(t, true, "restore")
		expect(t, true, "another")
	})

	t.Run("allowed-is-not-denied", func(t *testing.T) {
		reset()
		p.AllowedCommands = []string{"backup", "restore", "repair*"}
		p.DeniedCommands = []string{"backup", "repair-snapshot"}
		expect(t, true, "backup")
		expect(t, true, "restore")
		expect(t, true, "repair-index")
		expect(t, false, "repair-snapshot") // direct denied match, wildcard allowed results in denied
		expect(t, false, "another")
	})
}

func TestShowSchedules(t *testing.T) {
	buffer := &bytes.Buffer{}
	schedules := []*config.Schedule{
		{
			Profiles:    []string{"default"},
			CommandName: "check",
			Schedules:   []string{"weekly"},
		},
		{
			Profiles:    []string{"default"},
			CommandName: "backup",
			Schedules:   []string{"daily"},
		},
	}
	expected := strings.TrimSpace(`
schedule check@default:
    run:       check
    profiles:  default
    schedule:  weekly

schedule backup@default:
    run:       backup
    profiles:  default
    schedule:  daily

`)
	showSchedules(buffer, schedules)
	assert.Equal(t, expected, strings.TrimSpace(buffer.String()))
}

func TestCreateScheduleWhenNoneAvailable(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = createSchedule(nil, commandContext{
		Context: Context{
			config: cfg,
			flags: commandLineFlags{
				name: "default",
			},
		},
	})
	assert.Error(t, err)
}

func TestCreateScheduleAll(t *testing.T) {
	// loads an (almost) empty config
	// note that a default (or specific) profile is needed to load all schedules:
	// TODO: we should be able to load them all without a default profile
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = createSchedule(nil, commandContext{
		Context: Context{
			config: cfg,
			flags: commandLineFlags{
				name: "default",
			},
			request: Request{arguments: []string{"--all"}},
		},
	})
	assert.NoError(t, err)
}

func TestPreRunScheduleNoScheduleName(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = preRunSchedule(&Context{
		config: cfg,
		flags: commandLineFlags{
			name: "default",
		},
	})
	assert.Error(t, err)
	t.Log(err)
}

func TestPreRunScheduleWrongScheduleName(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = preRunSchedule(&Context{
		request: Request{arguments: []string{"wrong"}},
		config:  cfg,
		flags: commandLineFlags{
			name: "default",
		},
	})
	assert.Error(t, err)
	t.Log(err)
}

func TestPreRunScheduleProfileUnknown(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = preRunSchedule(&Context{
		request: Request{arguments: []string{"backup@profile"}},
		config:  cfg,
	})
	assert.ErrorIs(t, err, config.ErrNotFound)
}

func TestRunScheduleNoScheduleName(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = runSchedule(nil, commandContext{
		Context: Context{
			config: cfg,
			flags: commandLineFlags{
				name: "default",
			},
		},
	})
	assert.Error(t, err)
	t.Log(err)
}

func TestRunScheduleWrongScheduleName(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = runSchedule(nil, commandContext{
		Context: Context{
			request: Request{arguments: []string{"wrong"}},
			config:  cfg,
			flags: commandLineFlags{
				name: "default",
			},
		},
	})
	assert.Error(t, err)
	t.Log(err)
}

func TestRunScheduleProfileUnknown(t *testing.T) {
	// loads an (almost) empty config
	cfg, err := config.Load(bytes.NewBufferString("[default]"), "toml")
	assert.NoError(t, err)

	err = runSchedule(nil, commandContext{
		Context: Context{
			request: Request{arguments: []string{"backup@profile"}},
			config:  cfg,
		},
	})
	assert.ErrorIs(t, err, ErrProfileNotFound)
}
