package main

import (
	"fmt"
	"os"
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
		dryRun:          setEnv(true, "RESTICPROFILE_DRY_RUN").(bool),
		noLock:          setEnv(true, "RESTICPROFILE_NO_LOCK").(bool),
		lockWait:        setEnv(time.Minute*5, "RESTICPROFILE_LOCK_WAIT").(time.Duration),
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

func TestProfileCommandWithProfileNamePrecedence(t *testing.T) {
	_, flags, err := loadFlags([]string{"-n", "profile2", "-v", "some.command"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, "profile2")
	assert.True(t, flags.verbose)
	assert.Equal(t, flags.resticArgs, []string{"some.command"})
}

func TestProfileCommandWithProfileNamePrecedenceWithDefaultProfile(t *testing.T) {
	_, flags, err := loadFlags([]string{"-n", constants.DefaultProfileName, "-v", "some.other.command"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, constants.DefaultProfileName)
	assert.True(t, flags.verbose)
	assert.Equal(t, flags.resticArgs, []string{"some.other.command"})
}

func TestProfileCommandWithResticVerbose(t *testing.T) {
	_, flags, err := loadFlags([]string{"profile1.check", "--", "-v"})
	require.NoError(t, err)
	assert.False(t, flags.help)
	assert.Equal(t, flags.name, "profile1")
	assert.False(t, flags.verbose)
	assert.Equal(t, flags.resticArgs, []string{"check", "--", "-v"})
}

func TestProfileCommandWithResticprofileVerbose(t *testing.T) {
	_, flags, err := loadFlags([]string{"-v", "pro.file1.backup"})
	require.NoError(t, err)
	assert.True(t, flags.verbose)
	assert.Equal(t, flags.name, "pro.file1")
	assert.Equal(t, flags.resticArgs, []string{"backup"})
}

func TestProfileCommandThreePart(t *testing.T) {
	_, flags, err := loadFlags([]string{"bla.foo.backup"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, "bla.foo")
	assert.Equal(t, flags.resticArgs, []string{"backup"})
}

func TestProfileCommandTwoPartCommandMissing(t *testing.T) {
	_, flags, err := loadFlags([]string{"bar."})
	require.NoError(t, err)
	assert.Equal(t, flags.name, "bar")
	assert.Equal(t, flags.resticArgs, []string{})
}

func TestProfileCommandTwoPartProfileMissing(t *testing.T) {
	_, flags, err := loadFlags([]string{".baz"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, constants.DefaultProfileName)
	assert.Equal(t, flags.resticArgs, []string{"baz"})
}

func TestProfileCommandTwoPartDotPrefix(t *testing.T) {
	_, flags, err := loadFlags([]string{".baz.qux"})
	require.NoError(t, err)
	assert.Equal(t, flags.name, ".baz")
	assert.Equal(t, flags.resticArgs, []string{"qux"})
}

func TestProfileCommandThreePartCommandMissing(t *testing.T) {
	_, flags, err := loadFlags([]string{"quux.quuz."})
	require.NoError(t, err)
	assert.Equal(t, flags.name, "quux.quuz")
	assert.Equal(t, flags.resticArgs, []string{})
}

func TestProfileCommandTwoPartDotPrefixCommandMissing(t *testing.T) {
	_, flags, err := loadFlags([]string{".corge."})
	require.NoError(t, err)
	assert.Equal(t, flags.name, ".corge")
	assert.Equal(t, flags.resticArgs, []string{})
}

func TestProfileCommandTwoPartProfileMissingCommandMissing(t *testing.T) {
	_, flags, err := loadFlags([]string{"."})
	require.NoError(t, err)
	assert.Equal(t, flags.name, constants.DefaultProfileName)
	assert.Equal(t, flags.resticArgs, []string{})
}
