package schedule_test

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/schedule/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateJobHappyPath(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{}).Return(nil)
	handler.EXPECT().ParseSchedules([]string{}).Return([]*calendar.Event{}, nil)
	handler.EXPECT().CreateJob(mock.AnythingOfType("*schedule.Config"), []*calendar.Event{}, "user").Return(nil)

	scheduler := schedule.NewScheduler(handler, "profile")
	job := scheduler.NewJob(&schedule.Config{
		ProfileName: "profile",
		CommandName: "backup",
		Schedules:   []string{},
		Permission:  "user",
	})

	err := job.Create()
	require.NoError(t, err)
}

func TestCreateJobErrorParseSchedules(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{}).Return(nil)
	handler.EXPECT().ParseSchedules([]string{}).Return(nil, errors.New("test!"))

	scheduler := schedule.NewScheduler(handler, "profile")
	job := scheduler.NewJob(&schedule.Config{
		ProfileName: "profile",
		CommandName: "backup",
		Schedules:   []string{},
	})

	err := job.Create()
	require.Error(t, err)
}

func TestCreateJobErrorDisplaySchedules(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{}).Return(errors.New("test!"))

	scheduler := schedule.NewScheduler(handler, "profile")
	job := scheduler.NewJob(&schedule.Config{
		ProfileName: "profile",
		CommandName: "backup",
		Schedules:   []string{},
	})

	err := job.Create()
	require.Error(t, err)
}

func TestCreateJobErrorCreate(t *testing.T) {
	handler := mocks.NewHandler(t)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{}).Return(nil)
	handler.EXPECT().ParseSchedules([]string{}).Return([]*calendar.Event{}, nil)
	handler.EXPECT().CreateJob(mock.AnythingOfType("*schedule.Config"), []*calendar.Event{}, "user").Return(errors.New("test!"))

	scheduler := schedule.NewScheduler(handler, "profile")
	job := scheduler.NewJob(&schedule.Config{
		ProfileName: "profile",
		CommandName: "backup",
		Schedules:   []string{},
		Permission:  "user",
	})

	err := job.Create()
	require.Error(t, err)
}
