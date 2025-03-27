package main

import (
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/schedule/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//nolint:unparam
func configForJob(command string, at ...string) *config.Schedule {
	origin := config.ScheduleOrigin("profile", command)
	return config.NewDefaultSchedule(nil, origin, at...)
}

func TestScheduleNilJobs(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	err := scheduleJobs(handler, nil)
	assert.NoError(t, err)
}

func TestSimpleScheduleJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().ParseSchedules([]string{"sched"}).Return([]*calendar.Event{{}}, nil)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{"sched"}).Return(nil)
	handler.EXPECT().CreateJob(
		mock.AnythingOfType("*schedule.Config"),
		mock.AnythingOfType("[]*calendar.Event"),
		schedule.PermissionUserBackground).
		RunAndReturn(func(scheduleConfig *schedule.Config, events []*calendar.Event, permission schedule.Permission) error {
			assert.Equal(t, []string{"--no-ansi", "--config", `config file`, "run-schedule", "backup@profile"}, scheduleConfig.Arguments.RawArgs())
			assert.Equal(t, `--no-ansi --config "config file" run-schedule backup@profile`, scheduleConfig.Arguments.String())
			return nil
		})

	scheduleConfig := configForJob("backup", "sched")
	scheduleConfig.ConfigFile = "config file"
	err := scheduleJobs(handler, []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestFailScheduleJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().ParseSchedules([]string{"sched"}).Return([]*calendar.Event{{}}, nil)
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{"sched"}).Return(nil)
	handler.EXPECT().CreateJob(
		mock.AnythingOfType("*schedule.Config"),
		mock.AnythingOfType("[]*calendar.Event"),
		schedule.PermissionUserBackground).
		Return(errors.New("error creating job"))

	scheduleConfig := configForJob("backup", "sched")
	err := scheduleJobs(handler, []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}

func TestRemoveNilJobs(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	err := removeJobs(handler, nil)
	assert.NoError(t, err)
}

func TestRemoveJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), schedule.PermissionUserBackground).
		RunAndReturn(func(scheduleConfig *schedule.Config, _ schedule.Permission) error {
			assert.Equal(t, "profile", scheduleConfig.ProfileName)
			assert.Equal(t, "backup", scheduleConfig.CommandName)
			return nil
		})

	scheduleConfig := configForJob("backup", "sched")
	err := removeJobs(handler, []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestRemoveJobNoConfig(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), schedule.PermissionUserBackground).
		RunAndReturn(func(scheduleConfig *schedule.Config, _ schedule.Permission) error {
			assert.Equal(t, "profile", scheduleConfig.ProfileName)
			assert.Equal(t, "backup", scheduleConfig.CommandName)
			return nil
		})

	scheduleConfig := configForJob("backup")
	err := removeJobs(handler, []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestFailRemoveJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), schedule.PermissionUserBackground).
		Return(errors.New("error removing job"))

	scheduleConfig := configForJob("backup", "sched")
	err := removeJobs(handler, []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}

func TestNoFailRemoveUnknownJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), schedule.PermissionUserBackground).
		Return(schedule.ErrScheduledJobNotFound)

	scheduleConfig := configForJob("backup", "sched")
	err := removeJobs(handler, []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestNoFailRemoveUnknownRemoveOnlyJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionAuto).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)
	handler.EXPECT().RemoveJob(mock.AnythingOfType("*schedule.Config"), schedule.PermissionUserBackground).
		Return(schedule.ErrScheduledJobNotFound)

	scheduleConfig := configForJob("backup")
	err := removeJobs(handler, []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestStatusNilJobs(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DisplayStatus("profile").Return(nil)

	err := statusJobs(handler, "profile", nil)
	assert.NoError(t, err)
}

func TestStatusJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()
	handler.EXPECT().DisplaySchedules("profile", "backup", []string{"sched"}).Return(nil)
	handler.EXPECT().DisplayJobStatus(mock.AnythingOfType("*schedule.Config")).Return(nil)
	handler.EXPECT().DisplayStatus("profile").Return(nil)

	scheduleConfig := configForJob("backup", "sched")
	err := statusJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.NoError(t, err)
}

func TestStatusRemoveOnlyJob(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	scheduleConfig := configForJob("backup")
	err := statusJobs(handler, "profile", []*config.Schedule{scheduleConfig})
	assert.Error(t, err)
}

func TestRemoveScheduledJobs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		removeProfileName string
		fromConfigFile    string
		scheduledConfigs  []schedule.Config
		removedConfigs    []schedule.Config
		permission        schedule.Permission
	}{
		{
			removeProfileName: "profile_no_config",
			fromConfigFile:    "configFile",
			scheduledConfigs:  []schedule.Config{},
			removedConfigs:    []schedule.Config{},
			permission:        schedule.PermissionUserBackground,
		},
		{
			removeProfileName: "profile_one_config_to_remove",
			fromConfigFile:    "configFile",
			scheduledConfigs: []schedule.Config{
				{
					ProfileName: "profile_one_config_to_remove",
					CommandName: "backup",
					ConfigFile:  "configFile",
					Permission:  constants.SchedulePermissionUser,
				},
			},
			removedConfigs: []schedule.Config{
				{
					ProfileName: "profile_one_config_to_remove",
					CommandName: "backup",
					ConfigFile:  "configFile",
					Permission:  constants.SchedulePermissionUser,
				},
			},
			permission: schedule.PermissionUserBackground,
		},
		{
			removeProfileName: "profile_different_config_file",
			fromConfigFile:    "configFile",
			scheduledConfigs: []schedule.Config{
				{
					ProfileName: "profile_different_config_file",
					CommandName: "backup",
					ConfigFile:  "other_configFile",
					Permission:  constants.SchedulePermissionUser,
				},
			},
			removedConfigs: []schedule.Config{},
			permission:     schedule.PermissionUserBackground,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.removeProfileName, func(t *testing.T) {
			handler := mocks.NewHandler(t)
			handler.EXPECT().Init().Return(nil)
			handler.EXPECT().Close()

			handler.EXPECT().Scheduled(tc.removeProfileName).Return(tc.scheduledConfigs, nil)
			for _, cfg := range tc.removedConfigs {
				handler.EXPECT().RemoveJob(&cfg, tc.permission).Return(nil)
				handler.EXPECT().DetectSchedulePermission(tc.permission).Return(tc.permission, true)
				handler.EXPECT().CheckPermission(mock.Anything, tc.permission).Return(true)
			}

			err := removeScheduledJobs(handler, tc.fromConfigFile, tc.removeProfileName)
			assert.NoError(t, err)
		})
	}
}

func TestFailRemoveScheduledJobs(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	handler.EXPECT().Scheduled("profile_to_remove").Return([]schedule.Config{
		{
			ProfileName: "profile_to_remove",
			CommandName: "backup",
			ConfigFile:  "configFile",
			Permission:  constants.SchedulePermissionUser,
		},
	}, nil)
	handler.EXPECT().RemoveJob(&schedule.Config{
		ProfileName: "profile_to_remove",
		CommandName: "backup",
		ConfigFile:  "configFile",
		Permission:  constants.SchedulePermissionUser,
	}, schedule.PermissionUserBackground).Return(errors.New("impossible"))
	handler.EXPECT().DetectSchedulePermission(schedule.PermissionUserBackground).Return(schedule.PermissionUserBackground, true)
	handler.EXPECT().CheckPermission(mock.Anything, schedule.PermissionUserBackground).Return(true)

	err := removeScheduledJobs(handler, "configFile", "profile_to_remove")
	assert.Error(t, err)
	t.Log(err)
}

func TestStatusScheduledJobs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		statusProfileName string
		fromConfigFile    string
		scheduledConfigs  []schedule.Config
		statusConfigs     []schedule.Config
	}{
		{
			statusProfileName: "profile_no_config",
			fromConfigFile:    "configFile",
			scheduledConfigs:  []schedule.Config{},
			statusConfigs:     []schedule.Config{},
		},
		{
			statusProfileName: "profile_one_config_to_remove",
			fromConfigFile:    "configFile",
			scheduledConfigs: []schedule.Config{
				{
					ProfileName: "profile_one_config_to_remove",
					CommandName: "backup",
					ConfigFile:  "configFile",
				},
			},
			statusConfigs: []schedule.Config{
				{
					ProfileName: "profile_one_config_to_remove",
					CommandName: "backup",
					ConfigFile:  "configFile",
				},
			},
		},
		{
			statusProfileName: "profile_different_config_file",
			fromConfigFile:    "configFile",
			scheduledConfigs: []schedule.Config{
				{
					ProfileName: "profile_different_config_file",
					CommandName: "backup",
					ConfigFile:  "other_configFile",
				},
			},
			statusConfigs: []schedule.Config{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.statusProfileName, func(t *testing.T) {
			handler := mocks.NewHandler(t)
			handler.EXPECT().Init().Return(nil)
			handler.EXPECT().Close()

			handler.EXPECT().Scheduled(tc.statusProfileName).Return(tc.scheduledConfigs, nil)
			for _, cfg := range tc.statusConfigs {
				handler.EXPECT().DisplaySchedules(cfg.ProfileName, cfg.CommandName, []string(nil)).Return(nil)
				handler.EXPECT().DisplayJobStatus(&cfg).Return(nil)
			}
			if len(tc.scheduledConfigs) > 0 {
				handler.EXPECT().DisplayStatus(tc.statusProfileName).Return(nil)
			}

			err := statusScheduledJobs(handler, tc.fromConfigFile, tc.statusProfileName)
			assert.NoError(t, err)
		})
	}
}

func TestFailStatusScheduledJobs(t *testing.T) {
	t.Parallel()

	handler := mocks.NewHandler(t)
	handler.EXPECT().Init().Return(nil)
	handler.EXPECT().Close()

	handler.EXPECT().Scheduled("profile_name").Return([]schedule.Config{
		{
			ProfileName: "profile_name",
			CommandName: "backup",
			ConfigFile:  "configFile",
			Permission:  constants.SchedulePermissionUser,
		},
	}, nil)
	handler.EXPECT().DisplaySchedules("profile_name", "backup", []string(nil)).Return(errors.New("impossible"))
	handler.EXPECT().DisplayStatus("profile_name").Return(errors.New("impossible"))

	err := statusScheduledJobs(handler, "configFile", "profile_name")
	assert.Error(t, err)
	t.Log(err)
}
