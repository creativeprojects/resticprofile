package main

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/schedule/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScheduleNilJobs(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	err := scheduleJobs(handler, "profile", nil)
	assert.NoError(t, err)
}

func TestSimpleScheduleJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().ParseSchedules([]string{"sched"}).Return([]*calendar.Event{{}}, nil)
	handler.EXPECT().DisplayParsedSchedules("backup", []*calendar.Event{{}})
	handler.EXPECT().CreateJob(
		mock.AnythingOfType("*schedule.Config"),
		mock.AnythingOfType("[]*calendar.Event"),
		mock.AnythingOfType("string")).
		RunAndReturn(func(scheduleConfig *schedule.Config, events []*calendar.Event, permission string) error {
			assert.Equal(t, []string{"--no-ansi", "--config", "", "run-schedule", "backup@profile"}, scheduleConfig.Arguments)
			return nil
		})

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := scheduleJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestFailScheduleJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().ParseSchedules([]string{"sched"}).Return([]*calendar.Event{{}}, nil)
	handler.EXPECT().DisplayParsedSchedules("backup", []*calendar.Event{{}})
	handler.EXPECT().CreateJob(
		mock.AnythingOfType("*schedule.Config"),
		mock.AnythingOfType("[]*calendar.Event"),
		mock.AnythingOfType("string")).
		Return(errors.New("error creating job"))

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := scheduleJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}

func TestRemoveNilJobs(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	err := removeJobs(handler, "profile", nil)
	assert.NoError(t, err)
}

func TestRemoveJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), mock.AnythingOfType("string")).
		RunAndReturn(func(scheduleConfig *schedule.Config, user string) error {
			assert.Equal(t, "profile", scheduleConfig.ProfileName)
			assert.Equal(t, "backup", scheduleConfig.CommandName)
			return nil
		})

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := removeJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestRemoveJobNoConfig(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), mock.AnythingOfType("string")).
		RunAndReturn(func(scheduleConfig *schedule.Config, user string) error {
			assert.Equal(t, "profile", scheduleConfig.ProfileName)
			assert.Equal(t, "backup", scheduleConfig.CommandName)
			return nil
		})

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
	}
	err := removeJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestFailRemoveJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), mock.AnythingOfType("string")).
		Return(errors.New("error removing job"))

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := removeJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}

func TestNoFailRemoveUnknownJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), mock.AnythingOfType("string")).
		Return(schedule.ErrorServiceNotFound)

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := removeJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestNoFailRemoveUnknownRemoveOnlyJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), mock.AnythingOfType("string")).
		Return(schedule.ErrorServiceNotFound)

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
	}
	err := removeJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestStatusNilJobs(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DisplayStatus("profile").Return(nil)

	err := statusJobs(handler, "profile", nil)
	assert.NoError(t, err)
}

func TestStatusJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().ParseSchedules([]string{"sched"}).Return([]*calendar.Event{{}}, nil)
	handler.EXPECT().DisplayParsedSchedules("backup", []*calendar.Event{{}})
	handler.EXPECT().DisplayJobStatus(mock.AnythingOfType("*schedule.Config")).Return(nil)
	handler.EXPECT().DisplayStatus("profile").Return(nil)

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
		Schedules:   []string{"sched"},
	}
	err := statusJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestStatusRemoveOnlyJob(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	scheduleConfig := &config.Schedule{
		Profiles:    []string{"profile"},
		CommandName: "backup",
	}
	err := statusJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}
