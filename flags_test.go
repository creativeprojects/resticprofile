package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/creativeprojects/resticprofile/constants"
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
