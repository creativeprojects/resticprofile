package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnknownShortFlag(t *testing.T) {
	_, _, err := loadFlags([]string{"-z"})
	require.Error(t, err)
}

func TestUnknownLongFlag(t *testing.T) {
	_, _, err := loadFlags([]string{"--does-not-exist"})
	require.Error(t, err)
}

func TestHelpShortFlag(t *testing.T) {
	_, flags, err := loadFlags([]string{"-h"})
	require.NoError(t, err)
	assert.True(t, flags.help)
}

func TestHelpLongFlag(t *testing.T) {
	_, flags, err := loadFlags([]string{"--help"})
	require.NoError(t, err)
	assert.True(t, flags.help)
}

func TestBackupProfileName(t *testing.T) {
	_, flags, err := loadFlags([]string{"-n", "profile1", "backup", "-v"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, "profile1")
	assert.False(t, flags.verbose)
	assert.Equal(t, flags.resticArgs, []string{"backup", "-v"})
}

func TestEnvOverrides(t *testing.T) {
	var envNames []string
	t.Cleanup(func() {
		for _, name := range envNames {
			_ = os.Unsetenv(name)
		}
	})

	setEnv := func(value any, key string) any {
		envNames = append(envNames, key)
		assert.NoError(t, os.Setenv(key, fmt.Sprintf("%v", value)))
		return value
	}

	flags := commandLineFlags{
		help:            false,
		quiet:           setEnv(true, "RESTICPROFILE_QUIET").(bool),
		verbose:         setEnv(true, "RESTICPROFILE_VERBOSE").(bool),
		veryVerbose:     setEnv(true, "RESTICPROFILE_TRACE").(bool),
		config:          setEnv("custom-conf", "RESTICPROFILE_CONFIG").(string),
		format:          setEnv("custom-format", "RESTICPROFILE_FORMAT").(string),
		name:            setEnv("custom-profile", "RESTICPROFILE_NAME").(string),
		log:             setEnv("custom.log", "RESTICPROFILE_LOG").(string),
		commandOutput:   setEnv("log", "RESTICPROFILE_COMMAND_OUTPUT").(string),
		dryRun:          setEnv(true, "RESTICPROFILE_DRY_RUN").(bool),
		noLock:          setEnv(true, "RESTICPROFILE_NO_LOCK").(bool),
		lockWait:        setEnv(time.Minute*5, "RESTICPROFILE_LOCK_WAIT").(time.Duration),
		stderr:          setEnv(true, "RESTICPROFILE_STDERR").(bool),
		noAnsi:          setEnv(true, "RESTICPROFILE_NO_ANSI").(bool),
		theme:           setEnv("custom-theme", "RESTICPROFILE_THEME").(string),
		noPriority:      setEnv(true, "RESTICPROFILE_NO_PRIORITY").(bool),
		wait:            setEnv(true, "RESTICPROFILE_WAIT").(bool),
		ignoreOnBattery: setEnv(50, "RESTICPROFILE_IGNORE_ON_BATTERY").(int),
	}

	load := func(t *testing.T, args ...string) commandLineFlags {
		_, loaded, err := loadFlags(args)
		assert.NoError(t, err)
		flags.resticArgs = loaded.resticArgs
		flags.usagesHelp = loaded.usagesHelp
		return loaded
	}

	t.Run("all-defined-by-env", func(t *testing.T) {
		loaded := load(t, "--verbose")
		assert.Equal(t, flags, loaded)
	})

	t.Run("cli-has-higher-prio", func(t *testing.T) {
		loaded := load(t, "--name", "cli-profile-name")
		assert.NotEqual(t, flags, loaded)
		assert.Equal(t, "cli-profile-name", loaded.name)

		loaded.name = flags.name
		assert.Equal(t, flags, loaded)
	})
}

func TestEnvOverridesError(t *testing.T) {
	logger := clog.GetDefaultLogger()
	defer clog.SetDefaultLogger(logger)
	mem := clog.NewMemoryHandler()
	clog.SetDefaultLogger(clog.NewLogger(mem))

	assert.NoError(t, os.Setenv("RESTICPROFILE_LOCK_WAIT", "no-valid-duration"))
	t.Cleanup(func() { _ = os.Unsetenv("RESTICPROFILE_LOCK_WAIT") })

	_, loaded, err := loadFlags([]string{"--verbose"})
	assert.NoError(t, err)
	assert.NotNil(t, loaded)
	assert.Contains(t, mem.Logs(), `cannot convert env variable RESTICPROFILE_LOCK_WAIT="no-valid-duration": time: invalid duration "no-valid-duration"`)
}

func TestCommandProfileOrProfileCommandShortcuts(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		sourceArgs      []string
		expectedProfile string
		expectedArgs    []string
		expectedVerbose bool
	}{
		{
			sourceArgs:      []string{"."},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{".command"},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"profile."},
			expectedProfile: "profile",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{".profile."},
			expectedProfile: ".profile",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{"pro.file."},
			expectedProfile: "pro.file",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{".profile.command"},
			expectedProfile: ".profile",
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"pro.file.command"},
			expectedProfile: "pro.file",
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"-v", "pro.file1.backup"},
			expectedProfile: "pro.file1",
			expectedArgs:    []string{"backup"},
			expectedVerbose: true,
		},
		{
			sourceArgs:      []string{"profile1.check", "--", "-v"},
			expectedProfile: "profile1",
			expectedArgs:    []string{"check", "--", "-v"},
			expectedVerbose: false,
		},
		{
			sourceArgs:      []string{"-n", constants.DefaultProfileName, "-v", "some.other.command"},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{"some.other.command"},
			expectedVerbose: true,
		},
		{
			sourceArgs:      []string{"-n", "profile2", "-v", "some.command"},
			expectedProfile: "profile2",
			expectedArgs:    []string{"some.command"},
			expectedVerbose: true,
		},
		{
			sourceArgs:      []string{"@"},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{"command@"},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"@profile"},
			expectedProfile: "profile",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{"@profile@"},
			expectedProfile: "profile@",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{"command@profile@"},
			expectedProfile: "profile@",
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"@pro@file"},
			expectedProfile: "pro@file",
			expectedArgs:    []string{},
		},
		{
			sourceArgs:      []string{"command@pro@file"},
			expectedProfile: "pro@file",
			expectedArgs:    []string{"command"},
		},
		{
			sourceArgs:      []string{"-v", "backup@pro@file1"},
			expectedProfile: "pro@file1",
			expectedArgs:    []string{"backup"},
			expectedVerbose: true,
		},
		{
			sourceArgs:      []string{"check@profile1", "--", "-v"},
			expectedProfile: "profile1",
			expectedArgs:    []string{"check", "--", "-v"},
			expectedVerbose: false,
		},
		{
			sourceArgs:      []string{"-n", constants.DefaultProfileName, "-v", "some@other@command"},
			expectedProfile: constants.DefaultProfileName,
			expectedArgs:    []string{"some@other@command"},
			expectedVerbose: true,
		},
		{
			sourceArgs:      []string{"-n", "profile2", "-v", "some@command"},
			expectedProfile: "profile2",
			expectedArgs:    []string{"some@command"},
			expectedVerbose: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(strings.Join(testCase.sourceArgs, " "), func(t *testing.T) {
			t.Parallel()
			_, flags, err := loadFlags(testCase.sourceArgs)
			require.NoError(t, err)
			assert.Equal(t, flags.name, testCase.expectedProfile)
			assert.Equal(t, flags.resticArgs, testCase.expectedArgs)
			assert.Equal(t, flags.verbose, testCase.expectedVerbose)
		})
	}
}
