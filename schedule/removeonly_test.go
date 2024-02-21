package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	scheduler := NewScheduler(NewHandler(&SchedulerDefaultOS{}), profile)
	defer scheduler.Close()

	job := scheduler.NewJob(NewRemoveOnlyConfig(profile, "check"))

	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Create())
	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Status())
	assert.True(t, job.Accessible())
	assert.True(t, job.RemoveOnly())
	assert.NotEqual(t, ErrorJobCanBeRemovedOnly, job.Remove())
}
