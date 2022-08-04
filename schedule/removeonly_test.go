package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"

	"github.com/stretchr/testify/assert"
)

func TestNewRemoveOnlyConfig(t *testing.T) {
	c := config.NewRemoveOnlyConfig("profile", "command")

	assert.Equal(t, "profile", c.Title)
	assert.Equal(t, "command", c.SubTitle)
	assert.Equal(t, "", c.JobDescription)
	assert.Equal(t, "", c.TimerDescription)
	assert.Empty(t, c.Schedules)
	assert.Equal(t, "", c.Permission)
	assert.Equal(t, "", c.WorkingDirectory)
	assert.Equal(t, "", c.Command)
	assert.Empty(t, c.Arguments)
	assert.Empty(t, c.Environment)
	assert.Equal(t, "", c.Priority)
	assert.Equal(t, "", c.Log)
	assert.Equal(t, "", c.ConfigFile)
	{
		flag, found := c.GetFlag("")
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
