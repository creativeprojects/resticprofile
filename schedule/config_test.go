package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduleProperties(t *testing.T) {
	schedule := Config{
		ProfileName:             "profile",
		CommandName:             "command name",
		Schedules:               []string{"1", "2", "3"},
		Permission:              "admin",
		WorkingDirectory:        "home",
		Command:                 "command",
		Arguments:               NewCommandArguments([]string{"1", "2"}),
		Environment:             []string{"test=dev"},
		JobDescription:          "job",
		TimerDescription:        "timer",
		Priority:                "",
		ConfigFile:              "config",
		Flags:                   map[string]string{},
		removeOnly:              false,
		IgnoreOnBattery:         false,
		IgnoreOnBatteryLessThan: 0,
	}

	assert.Equal(t, "config", schedule.ConfigFile)
	assert.Equal(t, "profile", schedule.ProfileName)
	assert.Equal(t, "command name", schedule.CommandName)
	assert.Equal(t, "job", schedule.JobDescription)
	assert.Equal(t, "timer", schedule.TimerDescription)
	assert.ElementsMatch(t, []string{"1", "2", "3"}, schedule.Schedules)
	assert.Equal(t, "admin", schedule.Permission)
	assert.Equal(t, "command", schedule.Command)
	assert.Equal(t, "home", schedule.WorkingDirectory)
	assert.ElementsMatch(t, []string{"1", "2"}, schedule.Arguments.RawArgs())
	assert.Equal(t, []string{"test=dev"}, schedule.Environment)
	assert.Equal(t, "background", schedule.GetPriority()) // default value
}

func TestStandardPriority(t *testing.T) {
	schedule := Config{
		Priority: "standard",
	}
	assert.Equal(t, "standard", schedule.GetPriority())
}

func TestCaseInsensitivePriority(t *testing.T) {
	schedule := Config{
		Priority: "stANDard",
	}
	assert.Equal(t, "standard", schedule.GetPriority())
}

func TestOtherPriority(t *testing.T) {
	schedule := Config{
		Priority: "other",
	}
	assert.Equal(t, "background", schedule.GetPriority()) // default value
}

func TestScheduleFlags(t *testing.T) {
	schedule := &Config{}

	flag, found := schedule.GetFlag("unit")
	assert.Empty(t, flag)
	assert.False(t, found)

	schedule.SetFlag("unit", "test")
	flag, found = schedule.GetFlag("unit")
	assert.Equal(t, "test", flag)
	assert.True(t, found)
}
