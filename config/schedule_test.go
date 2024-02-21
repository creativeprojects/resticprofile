package config

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestScheduleProperties(t *testing.T) {
	schedule := Schedule{
		Profiles:                []string{"profile"},
		CommandName:             "command name",
		Schedules:               []string{"1", "2", "3"},
		Permission:              "admin",
		Environment:             []string{"test=dev"},
		Priority:                "",
		LockMode:                "undefined",
		LockWait:                1 * time.Minute,
		ConfigFile:              "config",
		Flags:                   map[string]string{},
		IgnoreOnBattery:         false,
		IgnoreOnBatteryLessThan: 0,
	}

	assert.Equal(t, "config", schedule.ConfigFile)
	assert.Equal(t, "profile", schedule.Profiles[0])
	assert.Equal(t, "command name", schedule.CommandName)
	assert.ElementsMatch(t, []string{"1", "2", "3"}, schedule.Schedules)
	assert.Equal(t, "admin", schedule.Permission)
	assert.Equal(t, []string{"test=dev"}, schedule.Environment)
	assert.Equal(t, ScheduleLockModeDefault, schedule.GetLockMode())
	assert.Equal(t, 60*time.Second, schedule.GetLockWait())
}

func TestLockModes(t *testing.T) {
	tests := map[ScheduleLockMode]Schedule{
		ScheduleLockModeDefault: {LockMode: ""},
		ScheduleLockModeFail:    {LockMode: constants.ScheduleLockModeOptionFail},
		ScheduleLockModeIgnore:  {LockMode: constants.ScheduleLockModeOptionIgnore},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.GetLockMode())
	}
}

func TestLockWait(t *testing.T) {
	tests := map[time.Duration]Schedule{
		0:               {LockWait: 2 * time.Second}, // min lock wait is is >2 seconds
		3 * time.Second: {LockWait: 3 * time.Second},
		120 * time.Hour: {LockWait: 120 * time.Hour},
	}
	for mode, config := range tests {
		assert.Equal(t, mode, config.GetLockWait())
	}
}

func TestScheduleFlags(t *testing.T) {
	schedule := &Schedule{}

	flag, found := schedule.GetFlag("unit")
	assert.Empty(t, flag)
	assert.False(t, found)

	schedule.SetFlag("unit", "test")
	flag, found = schedule.GetFlag("unit")
	assert.Equal(t, "test", flag)
	assert.True(t, found)
}
