//go:build !darwin && !windows

package schedule

import (
	"bytes"
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadingSystemdScheduled(t *testing.T) {
	event := calendar.NewEvent()
	require.NoError(t, event.Parse("2020-01-01"))

	schedulePermission := constants.SchedulePermissionUserLoggedOn

	testCases := []struct {
		job       Config
		schedules []*calendar.Event
	}{
		{
			job: Config{
				ProfileName:      "testscheduled",
				CommandName:      "backup",
				Command:          "/tmp/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Permission:       schedulePermission,
				ConfigFile:       "examples/dev.yaml",
				Schedules:        []string{event.String()},
			},
			schedules: []*calendar.Event{event},
		},
		{
			job: Config{
				ProfileName:      "test.scheduled",
				CommandName:      "backup",
				Command:          "/tmp/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "config file.yaml", "--name", "self", "backup"}),
				WorkingDirectory: "/resticprofile",
				Permission:       schedulePermission,
				ConfigFile:       "config file.yaml",
				Schedules:        []string{event.String()},
			},
			schedules: []*calendar.Event{event},
		},
	}
	userHome, err := os.UserHomeDir()
	require.NoError(t, err)

	handler := NewHandler(SchedulerSystemd{}).(*HandlerSystemd)

	expectedJobs := []Config{}
	for _, testCase := range testCases {
		job := testCase.job
		err := handler.CreateJob(&job, testCase.schedules, PermissionFromConfig(schedulePermission))

		toRemove := &job
		t.Cleanup(func() {
			_ = handler.RemoveJob(toRemove, PermissionFromConfig(schedulePermission))
		})
		require.NoError(t, err)

		job.Environment = []string{"HOME=" + userHome}
		expectedJobs = append(expectedJobs, job)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	testScheduled := make([]Config, 0, len(scheduled))
	for _, s := range scheduled {
		if s.ConfigFile != "config file.yaml" && s.ConfigFile != "examples/dev.yaml" {
			t.Logf("Ignoring config file %s", s.ConfigFile)
			continue
		}
		testScheduled = append(testScheduled, s)
	}

	assert.ElementsMatch(t, expectedJobs, testScheduled)

	// now delete all schedules
	for _, testCase := range testCases {
		err := handler.RemoveJob(&testCase.job, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err = handler.Scheduled("")
	require.NoError(t, err)
	assert.Empty(t, scheduled)
}

func TestDetectPermissionSystemd(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"", "user_logged_on", false},
		{"something", "user_logged_on", false},
		{"system", "system", true},
		{"user", "user", true},
		{"user_logged_on", "user_logged_on", true},
		{"user_logged_in", "user_logged_on", true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(SchedulerSystemd{}).(*HandlerSystemd)
			perm, safe := handler.DetectSchedulePermission(PermissionFromConfig(fixture.input))
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
func TestSystemdConfigPermission(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		config   systemd.Config
		expected string
	}{
		{
			name: "SystemUnit with User",
			config: systemd.Config{
				UnitType: systemd.SystemUnit,
				User:     "testuser",
			},
			expected: constants.SchedulePermissionUser,
		},
		{
			name: "SystemUnit without User",
			config: systemd.Config{
				UnitType: systemd.SystemUnit,
				User:     "",
			},
			expected: constants.SchedulePermissionSystem,
		},
		{
			name: "Default case (UserUnit)",
			config: systemd.Config{
				UnitType: systemd.UserUnit,
			},
			expected: constants.SchedulePermissionUserLoggedOn,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := systemdConfigPermission(testCase.config)
			assert.Equal(t, testCase.expected, result)
		})
	}
}
func TestPermissionToSystemd(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		permission   Permission
		isRoot       bool
		expected     systemd.UnitType
		expectedUser string
	}{
		{
			name:         "PermissionSystem",
			permission:   PermissionSystem,
			isRoot:       false,
			expected:     systemd.SystemUnit,
			expectedUser: "",
		},
		{
			name:         "PermissionUserBackground",
			permission:   PermissionUserBackground,
			isRoot:       false,
			expected:     systemd.SystemUnit,
			expectedUser: "testuser",
		},
		{
			name:         "PermissionUserLoggedOn",
			permission:   PermissionUserLoggedOn,
			isRoot:       false,
			expected:     systemd.UserUnit,
			expectedUser: "",
		},
		{
			name:         "Default case as non-root",
			permission:   PermissionFromConfig("unknown"),
			isRoot:       false,
			expected:     systemd.UserUnit,
			expectedUser: "",
		},
		{
			name:         "Default case as root",
			permission:   PermissionFromConfig("unknown"),
			isRoot:       true,
			expected:     systemd.SystemUnit,
			expectedUser: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			currentUser := user.User{
				Uid:      1000,
				Gid:      1000,
				Username: "testuser",
			}
			if testCase.isRoot {
				currentUser.Uid = 0
			}

			unitType, user := permissionToSystemd(currentUser, testCase.permission)
			assert.Equal(t, testCase.expected, unitType)
			assert.Equal(t, testCase.expectedUser, user)
		})
	}
}

func TestDisplaySystemdSchedulesWithEmpty(t *testing.T) {
	err := displaySystemdSchedules("profile", "command", []string{""})
	require.Error(t, err)
}

func TestDisplaySystemdSchedules(t *testing.T) {
	buffer := &bytes.Buffer{}
	term.SetOutput(buffer)
	defer term.SetOutput(os.Stdout)

	err := displaySystemdSchedules("profile", "command", []string{"daily"})
	require.NoError(t, err)

	output := buffer.String()
	assert.Contains(t, output, "Original form: daily")
	assert.Contains(t, output, "Normalized form: *-*-* 00:00:00")
}
