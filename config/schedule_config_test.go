package config

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestScheduleProperties(t *testing.T) {
	schedule := ScheduleConfig{
		ConfigFile:       "config",
		Title:            "profile",
		SubTitle:         "command name",
		Schedules:        []string{"1", "2", "3"},
		Permission:       "admin",
		WorkingDirectory: "home",
		Command:          "command",
		Arguments:        []string{"1", "2"},
		Environment:      map[string]string{"test": "dev"},
		JobDescription:   "job",
		TimerDescription: "timer",
		Log:              "log.txt",
		LockMode:         "undefined",
		LockWait:         1 * time.Minute,
	}

	assert.Equal(t, "config", schedule.ConfigFile)
	assert.Equal(t, "profile", schedule.Title)
	assert.Equal(t, "command name", schedule.SubTitle)
	assert.Equal(t, "job", schedule.JobDescription)
	assert.Equal(t, "timer", schedule.TimerDescription)
	assert.ElementsMatch(t, []string{"1", "2", "3"}, schedule.Schedules)
	assert.Equal(t, "admin", schedule.Permission)
	assert.Equal(t, "command", schedule.Command)
	assert.Equal(t, "home", schedule.WorkingDirectory)
	assert.ElementsMatch(t, []string{"1", "2"}, schedule.Arguments)
	assert.Equal(t, "dev", schedule.Environment["test"])
	assert.Equal(t, "background", schedule.GetPriority()) // default value
	assert.Equal(t, "log.txt", schedule.Log)
	assert.Equal(t, ScheduleLockModeDefault, schedule.GetLockMode())
	assert.Equal(t, 60*time.Second, schedule.GetLockWait())
}

func TestLockModes(t *testing.T) {
	tests := map[ScheduleLockMode]ScheduleConfig{
		ScheduleLockModeDefault: {LockMode: ""},
		ScheduleLockModeFail:    {LockMode: constants.ScheduleLockModeOptionFail},
		ScheduleLockModeIgnore:  {LockMode: constants.ScheduleLockModeOptionIgnore},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.GetLockMode())
	}
}

func TestLockWait(t *testing.T) {
	tests := map[time.Duration]ScheduleConfig{
		0:               {LockWait: 2 * time.Second}, // min lock wait is is >2 seconds
		3 * time.Second: {LockWait: 3 * time.Second},
		120 * time.Hour: {LockWait: 120 * time.Hour},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.GetLockWait())
	}
}

func TestStandardPriority(t *testing.T) {
	schedule := ScheduleConfig{
		Priority: "standard",
	}
	assert.Equal(t, "standard", schedule.GetPriority())
}

func TestCaseInsensitivePriority(t *testing.T) {
	schedule := ScheduleConfig{
		Priority: "stANDard",
	}
	assert.Equal(t, "standard", schedule.GetPriority())
}

func TestOtherPriority(t *testing.T) {
	schedule := ScheduleConfig{
		Priority: "other",
	}
	assert.Equal(t, "background", schedule.GetPriority()) // default value
}

func TestScheduleFlags(t *testing.T) {
	schedule := &ScheduleConfig{}

	flag, found := schedule.GetFlag("unit")
	assert.Empty(t, flag)
	assert.False(t, found)

	schedule.SetFlag("unit", "test")
	flag, found = schedule.GetFlag("unit")
	assert.Equal(t, "test", flag)
	assert.True(t, found)
}
