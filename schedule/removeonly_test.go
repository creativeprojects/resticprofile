package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRemoveOnlyConfig(t *testing.T) {
	cfg := NewRemoveOnlyConfig("profile", "command")

	assert.Equal(t, "profile", cfg.ProfileName)
	assert.Equal(t, "command", cfg.CommandName)
	assert.Equal(t, "", cfg.JobDescription)
	assert.Equal(t, "", cfg.TimerDescription)
	assert.Empty(t, cfg.Schedules)
	assert.Equal(t, "", cfg.Permission)
	assert.Equal(t, "", cfg.WorkingDirectory)
	assert.Equal(t, "", cfg.Command)
	assert.Empty(t, cfg.Arguments)
	assert.Empty(t, cfg.Environment)
	assert.Equal(t, "", cfg.Priority)
	assert.Equal(t, "", cfg.ConfigFile)
	{
		flag, found := cfg.GetFlag("")
		assert.Equal(t, "", flag)
		assert.False(t, found)
	}
}

func TestDetectRemoveOnlyConfig(t *testing.T) {
	assertRemoveOnly := func(expected bool, config *Config) {
		assert.Equal(t, expected, config.removeOnly)
	}

	assertRemoveOnly(true, NewRemoveOnlyConfig("", ""))
	assertRemoveOnly(false, &Config{})
}

func TestRemoveOnlyJob(t *testing.T) {
	profile := "non-existent"
	handler := NewHandler(SchedulerDefaultOS{})
	require.NoError(t, handler.Init())
	defer handler.Close()

	job := NewJob(handler, NewRemoveOnlyConfig(profile, "check"))

	assert.Equal(t, ErrJobCanBeRemovedOnly, job.Create())
	assert.Equal(t, ErrJobCanBeRemovedOnly, job.Status())
	assert.True(t, job.Accessible())
	assert.True(t, job.RemoveOnly())
	assert.NotEqual(t, ErrJobCanBeRemovedOnly, job.Remove())
}
