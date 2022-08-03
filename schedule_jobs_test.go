package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/mocks"
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

func TestScheduleEmptyJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.On("Init").Return(nil)
	handler.On("Close")
	handler.On("ParseSchedules", []string(nil)).Return(nil, nil)
	handler.On("DisplaySchedules", "", []string(nil)).Return(nil)
	handler.On("CreateJob", mock.AnythingOfType("*config.ScheduleConfig"), mock.AnythingOfType("[]*calendar.Event"), "user").Return(nil)

	scheduleConfig := &config.ScheduleConfig{}
	err := scheduleJobs(handler, "profile", []*config.ScheduleConfig{scheduleConfig})
	assert.NoError(t, err)
}
