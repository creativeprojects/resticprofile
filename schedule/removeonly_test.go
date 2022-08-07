package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"

	"github.com/stretchr/testify/assert"
)

func TestNewRemoveOnlyConfig(t *testing.T) {
	cfg := config.NewRemoveOnlyConfig("profile", "command")

	assert.Equal(t, "profile", cfg.Title)
	assert.Equal(t, "command", cfg.SubTitle)
	assert.Equal(t, "", cfg.JobDescription)
	assert.Equal(t, "", cfg.TimerDescription)
	assert.Empty(t, cfg.Schedules)
	assert.Equal(t, "", cfg.Permission)
	assert.Equal(t, "", cfg.WorkingDirectory)
	assert.Equal(t, "", cfg.Command)
	assert.Empty(t, cfg.Arguments)
	assert.Empty(t, cfg.Environment)
	assert.Equal(t, "", cfg.Priority)
	assert.Equal(t, "", cfg.Log)
	assert.Equal(t, "", cfg.ConfigFile)
	{
		flag, found := cfg.GetFlag("")
		assert.Equal(t, "", flag)
		assert.False(t, found)
	}
}

func TestDetectRemoveOnlyConfig(t *testing.T) {
	assertRemoveOnly := func(expected bool, config *config.ScheduleConfig) {
		assert.Equal(t, expected, config.RemoveOnly)
	}

	assertRemoveOnly(true, config.NewRemoveOnlyConfig("", ""))
	assertRemoveOnly(false, &config.ScheduleConfig{})
}

func TestRemoveOnlyJob(t *testing.T) {
	profile := "non-existent"
	scheduler := NewScheduler(NewHandler(&SchedulerDefaultOS{}), profile)
	defer scheduler.Close()

	job := scheduler.NewJob(config.NewRemoveOnlyConfig(profile, "check"))

	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Create())
	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Status())
	assert.True(t, job.Accessible())
	assert.True(t, job.RemoveOnly())
	assert.NotEqual(t, ErrorJobCanBeRemovedOnly, job.Remove())
}
