package schedule

import (
	"github.com/creativeprojects/resticprofile/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRemoveOnlyConfig(t *testing.T) {
	c := NewRemoveOnlyConfig("profile", "command")

	assert.Equal(t, "profile", c.Title())
	assert.Equal(t, "command", c.SubTitle())
	assert.Equal(t, "", c.JobDescription())
	assert.Equal(t, "", c.TimerDescription())
	assert.Empty(t, c.Schedules())
	assert.Equal(t, "", c.Permission())
	assert.Equal(t, "", c.WorkingDirectory())
	assert.Equal(t, "", c.Command())
	assert.Empty(t, c.Arguments())
	assert.Empty(t, c.Environment())
	assert.Equal(t, "", c.Priority())
	assert.Equal(t, "", c.Logfile())
	assert.Equal(t, "", c.Configfile())
	{
		flag, found := c.GetFlag("")
		assert.Equal(t, "", flag)
		assert.False(t, found)
	}
}

func TestDetectRemoveOnlyConfig(t *testing.T) {
	assertRemoveOnly := func(expected bool, config Config) {
		assert.Equal(t, expected, isRemoveOnlyConfig(config))
	}

	assertRemoveOnly(true, NewRemoveOnlyConfig("", ""))
	assertRemoveOnly(false, nil)
	assertRemoveOnly(false, &config.ScheduleConfig{})
}

func TestRemoveOnlyJob(t *testing.T) {
	profile := "non-existent"
	scheduler := NewScheduler("", profile)
	defer scheduler.Close()

	config := NewRemoveOnlyConfig(profile, "check")
	job := scheduler.NewJob(config)

	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Create())
	assert.Equal(t, ErrorJobCanBeRemovedOnly, job.Status())
	assert.True(t, job.Accessible())
	assert.True(t, job.RemoveOnly())
	assert.NotEqual(t, ErrorJobCanBeRemovedOnly, job.Remove())
}
