package config

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestScheduleProperties(t *testing.T) {
	schedule := ScheduleConfig{
		configfile:       "config",
		profileName:      "profile",
		commandName:      "command name",
		schedules:        []string{"1", "2", "3"},
		permission:       "admin",
		wd:               "home",
		command:          "command",
		arguments:        []string{"1", "2"},
		environment:      map[string]string{"test": "dev"},
		jobDescription:   "job",
		timerDescription: "timer",
		logfile:          "log.txt",
		lockMode:         "undefined",
		lockWait:         1 * time.Minute,
	}

	assert.Equal(t, "config", schedule.Configfile())
	assert.Equal(t, "profile", schedule.Title())
	assert.Equal(t, "command name", schedule.SubTitle())
	assert.Equal(t, "job", schedule.JobDescription())
	assert.Equal(t, "timer", schedule.TimerDescription())
	assert.ElementsMatch(t, []string{"1", "2", "3"}, schedule.Schedules())
	assert.Equal(t, "admin", schedule.Permission())
	assert.Equal(t, "command", schedule.Command())
	assert.Equal(t, "home", schedule.WorkingDirectory())
	assert.ElementsMatch(t, []string{"1", "2"}, schedule.Arguments())
	assert.Equal(t, "dev", schedule.Environment()["test"])
	assert.Equal(t, "background", schedule.Priority()) // default value
	assert.Equal(t, "log.txt", schedule.Logfile())
	assert.Equal(t, ScheduleLockModeDefault, schedule.LockMode())
	assert.Equal(t, 60*time.Second, schedule.LockWait())
}

func TestLockModes(t *testing.T) {
	tests := map[ScheduleLockMode]ScheduleConfig{
		ScheduleLockModeDefault: {lockMode: ""},
		ScheduleLockModeFail:    {lockMode: constants.ScheduleLockModeOptionFail},
		ScheduleLockModeIgnore:  {lockMode: constants.ScheduleLockModeOptionIgnore},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.LockMode())
	}
}

func TestLockWait(t *testing.T) {
	tests := map[time.Duration]ScheduleConfig{
		0:               {lockWait: 2 * time.Second}, // min lock wait is is >2 seconds
		3 * time.Second: {lockWait: 3 * time.Second},
		120 * time.Hour: {lockWait: 120 * time.Hour},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.LockWait())
	}
}

func TestStandardPriority(t *testing.T) {
	schedule := ScheduleConfig{
		priority: "standard",
	}
	assert.Equal(t, "standard", schedule.Priority())
}

func TestCaseInsensitivePriority(t *testing.T) {
	schedule := ScheduleConfig{
		priority: "stANDard",
	}
	assert.Equal(t, "standard", schedule.Priority())
}

func TestOtherPriority(t *testing.T) {
	schedule := ScheduleConfig{
		priority: "other",
	}
	assert.Equal(t, "background", schedule.Priority()) // default value
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
