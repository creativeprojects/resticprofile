package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/schedule/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleNilJobs(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.On("Init").Return(nil)
	handler.On("Close")

	err := scheduleJobs(handler, "profile", nil)
	assert.NoError(t, err)
}

func TestArgumentsOnScheduleJobNoLog(t *testing.T) {
	handler := getMockHandler(t)
	handler.On("CreateJob",
		mock.AnythingOfType("*config.ScheduleConfig"),
		mock.AnythingOfType("[]*calendar.Event"),
		mock.AnythingOfType("string")).
		Return(func(scheduleConfig *config.ScheduleConfig, events []*calendar.Event, permission string) error {
			assert.Equal(t, []string{"--no-ansi", "--config", "", "--name", "profile", "backup"}, scheduleConfig.Arguments)
			return nil
		})

	scheduleConfig := &config.ScheduleConfig{
		Title:    "profile",
		SubTitle: "backup",
	}
	err := scheduleJobs(handler, "profile", []*config.ScheduleConfig{scheduleConfig})
	assert.NoError(t, err)
}

func TestArgumentsOnScheduleJobLogFile(t *testing.T) {
	handler := getMockHandler(t)
	handler.EXPECT().CreateJob(
		mock.AnythingOfType("*config.ScheduleConfig"),
		mock.AnythingOfType("[]*calendar.Event"),
		mock.AnythingOfType("string")).
		Run(func(scheduleConfig *config.ScheduleConfig, events []*calendar.Event, permission string) {
			assert.Equal(t, []string{"--no-ansi", "--config", "", "--name", "profile", "--log", "/path/to/file", "backup"}, scheduleConfig.Arguments)
		}).
		Return(nil)

	scheduleConfig := &config.ScheduleConfig{
		Title:    "profile",
		SubTitle: "backup",
		Log:      "/path/to/file",
	}
	err := scheduleJobs(handler, "profile", []*config.ScheduleConfig{scheduleConfig})
	assert.NoError(t, err)
}

func TestArgumentsOnScheduleJobLogSyslog(t *testing.T) {
	if !platform.SupportsSyslog() {
		t.Skip("syslog is not supported")
	}
	handler := getMockHandler(t)
	handler.On("CreateJob",
		mock.AnythingOfType("*config.ScheduleConfig"),
		mock.AnythingOfType("[]*calendar.Event"),
		mock.AnythingOfType("string")).
		Return(func(scheduleConfig *config.ScheduleConfig, events []*calendar.Event, permission string) error {
			assert.Equal(t, []string{"--no-ansi", "--config", "", "--name", "profile", "--log", "tcp://localhost:123", "backup"}, scheduleConfig.Arguments)
			return nil
		})

	scheduleConfig := &config.ScheduleConfig{
		Title:    "profile",
		SubTitle: "backup",
		Log:      "tcp://localhost:123",
	}
	err := scheduleJobs(handler, "profile", []*config.ScheduleConfig{scheduleConfig})
	assert.NoError(t, err)
}

func getMockHandler(t *testing.T) *mocks.Handler {
	t.Helper()
	handler := mocks.NewHandler(t)
	handler.On("Init").Return(nil)
	handler.On("Close")
	handler.On("ParseSchedules", []string(nil)).Return(nil, nil)
	handler.On("DisplaySchedules", "backup", []string(nil)).Return(nil)
	return handler
}
