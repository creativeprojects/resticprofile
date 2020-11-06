package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScheduleProperties(t *testing.T) {
	schedule := ScheduleConfig{
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
		nice:             11,
		logfile:          "log.txt",
	}

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
	assert.Equal(t, 11, schedule.Nice())
	assert.Equal(t, "log.txt", schedule.Logfile())
}
