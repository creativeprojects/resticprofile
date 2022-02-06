package main

import (
	"testing"

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
	assert.False(t, flags.help)
	assert.Equal(t, flags.name, "profile1")
	assert.Equal(t, flags.resticArgs, []string{"backup", "-v"})
}
