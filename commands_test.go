package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	_, _, notFoundErr := getRemovableScheduleJobs(parsedConfig, "non-existent")
	assert.ErrorIs(t, notFoundErr, config.ErrNotFound)

	// Test that declared and declarable job configs are returned
	_, schedules, err := getRemovableScheduleJobs(parsedConfig, "default")
	assert.Nil(t, err)
	assert.NotEmpty(t, schedules)
	assert.Len(t, schedules, len(config.NewProfile(parsedConfig, "test").SchedulableCommands()))

	declaredCount := 0

	for _, jobConfig := range schedules {
		configOrigin := jobConfig.ScheduleOrigin()
		handler := schedule.NewHandler(schedule.SchedulerDefaultOS{})
		require.NoError(t, handler.Init())
		defer func(s schedule.Handler) { s.Close() }(handler) // Capture current ref to scheduler to be able to close it when function returns.

		if configOrigin.Command == constants.CommandCheck {
			assert.False(t, schedule.NewJob(handler, scheduleToConfig(jobConfig)).RemoveOnly())
			declaredCount++
		} else {
			assert.True(t, schedule.NewJob(handler, scheduleToConfig(jobConfig)).RemoveOnly())
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
	_, _, notFoundErr := getProfileScheduleJobs(cfg, "non-existent")
	assert.ErrorIs(t, notFoundErr, config.ErrNotFound)

	// Test that non-existent schedule causes no error at first
	{
		_, schedules, err := getProfileScheduleJobs(cfg, "other")
		assert.Nil(t, err)

		err = requireScheduleJobs(schedules, "other")
		assert.EqualError(t, err, "no schedule found for profile 'other'")
	}

	// Test that only declared job configs are returned
	{
		profile, schedules, err := getProfileScheduleJobs(cfg, "default")
		assert.Nil(t, err)

		err = requireScheduleJobs(schedules, "default")
		assert.Nil(t, err)

		assert.NotNil(t, profile)
		assert.NotEmpty(t, schedules)
		assert.Len(t, schedules, 1)
		assert.Equal(t, "check", schedules[0].ScheduleOrigin().Command)
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
	allGroups := []string{"others", "default"}
	allProfilesAndGroups := append(allProfiles, allGroups...)

	cfg, err := config.Load(bytes.NewBufferString(testConfig), "toml")
	assert.Nil(t, err)
	for _, p := range allProfiles {
		assert.True(t, cfg.HasProfile(p))
	}
	for _, g := range allGroups {
		assert.True(t, cfg.HasProfileGroup(g))
	}

	// Select --all
	assert.ElementsMatch(t, allProfilesAndGroups, selectProfilesAndGroups(cfg, "", []string{"--all"}))

	// Select profiles by name
	for _, p := range allProfiles {
		assert.ElementsMatch(t, []string{p}, selectProfilesAndGroups(cfg, p, nil))
	}

	// Select groups by name
	for _, g := range allGroups {
		assert.ElementsMatch(t, []string{g}, selectProfilesAndGroups(cfg, g, nil))
	}

	// Select non-existing profile or group
	assert.ElementsMatch(t, []string{"non-existing"}, selectProfilesAndGroups(cfg, "non-existing", nil))
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
		return commandContext{Context: Context{request: Request{arguments: args}}}
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
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--config-reference", "--to", t.TempDir()})))
		ref := buffer.String()
		assert.Contains(t, ref, "generating reference.gomd")
		assert.Contains(t, ref, "generating profile section")
		assert.Contains(t, ref, "generating nested section")
	})

	t.Run("--json-schema global", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "global"})))
		ref := buffer.String()
		assert.Contains(t, ref, `"$schema"`)
		assert.Contains(t, ref, "/jsonschema/config-1.json")
		assert.Contains(t, ref, "/jsonschema/config-2.json")

		decoder := json.NewDecoder(strings.NewReader(ref))
		content := make(map[string]any)
		assert.NoError(t, decoder.Decode(&content))
		assert.Contains(t, content, `$schema`)
	})

	t.Run("--json-schema no-option", func(t *testing.T) {
		buffer.Reset()
		assert.Error(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema"})))
	})

	t.Run("--json-schema invalid-option", func(t *testing.T) {
		buffer.Reset()
		assert.Error(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "_invalid_"})))
	})

	t.Run("--json-schema v1", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "v1"})))
		ref := buffer.String()
		assert.Contains(t, ref, "/jsonschema/config-1.json")
	})

	t.Run("--json-schema v2", func(t *testing.T) {
		buffer.Reset()
		assert.NoError(t, generateCommand(buffer, contextWithArguments([]string{"--json-schema", "v2"})))
		ref := buffer.String()
		assert.Contains(t, ref, "\"profiles\":")
		assert.Contains(t, ref, "/jsonschema/config-2.json")
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

func TestShowSchedules(t *testing.T) {
	buffer := &bytes.Buffer{}
	create := func(command string, at ...string) *config.Schedule {
		origin := config.ScheduleOrigin("default", command)
		return config.NewDefaultSchedule(nil, origin, at...)
	}
	schedules := []*config.Schedule{
		create("check", "weekly"),
		create("backup", "daily"),
	}
	expected := strings.TrimSpace(`
schedule backup@default:
    at:                   daily
    permission:           auto
    run-level:            auto
    command-output:       auto
    priority:             standard
    lock-mode:            default
    capture-environment:  RESTIC_*

schedule check@default:
    at:                   weekly
    permission:           auto
    run-level:            auto
    command-output:       auto
    priority:             standard
    lock-mode:            default
    capture-environment:  RESTIC_*
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
			request: Request{
				profile:   "default",
				arguments: []string{"--all"},
			},
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

func TestBatteryCommand(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := batteryCommand(buffer, commandContext{})
	require.NoError(t, err)
}
