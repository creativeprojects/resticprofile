package preventsleep_test

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/preventsleep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaffeinate(t *testing.T) {
	caffeinate := preventsleep.New()
	assert.False(t, caffeinate.IsRunning())

	err := caffeinate.Start()
	if errors.Is(err, preventsleep.ErrPermissionDenied) {
		t.Skip("user must be root to run this test")
	}
	require.NoError(t, err)
	assert.True(t, caffeinate.IsRunning())

	err = caffeinate.Stop()
	require.NoError(t, err)
	assert.False(t, caffeinate.IsRunning())
}

func TestStopShouldReturnError(t *testing.T) {
	caffeinate := preventsleep.New()
	assert.False(t, caffeinate.IsRunning())

	err := caffeinate.Stop()
	assert.ErrorIs(t, err, preventsleep.ErrNotStarted)
}

func TestCannotStartTwice(t *testing.T) {
	caffeinate := preventsleep.New()
	assert.False(t, caffeinate.IsRunning())

	err := caffeinate.Start()
	if errors.Is(err, preventsleep.ErrPermissionDenied) {
		t.Skip("user must be root to run this test")
	}
	require.NoError(t, err)
	assert.True(t, caffeinate.IsRunning())

	err = caffeinate.Start()
	assert.ErrorIs(t, err, preventsleep.ErrAlreadyStarted)

	err = caffeinate.Stop()
	require.NoError(t, err)
}
