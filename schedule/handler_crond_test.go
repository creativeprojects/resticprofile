//go:build !windows

package schedule

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/crond"
	"github.com/creativeprojects/resticprofile/user"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerCrond(t *testing.T) {
	handler := NewHandler(SchedulerCrond{})
	assert.IsType(t, &HandlerCrond{}, handler)
}

func TestCreateReadDeleteCrondSchedules(t *testing.T) {
	hourly := calendar.NewEvent(func(e *calendar.Event) {
		e.Minute.MustAddValue(0)
		e.Second.MustAddValue(0)
	})

	testCases := []struct {
		job       Config
		schedules []*calendar.Event
	}{
		{
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "examples/dev.yaml",
				Permission:       "user",
			},
			schedules: []*calendar.Event{
				hourly,
			},
		},
		{
			job: Config{
				ProfileName:      "test.scheduled",
				CommandName:      "backup",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "config file.yaml", "--name", "test.scheduled", "backup"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "config file.yaml",
				Permission:       "user",
			},
			schedules: []*calendar.Event{
				hourly,
			},
		},
	}

	tempFile := filepath.Join(t.TempDir(), "crontab")
	handler := NewHandler(SchedulerCrond{
		CrontabFile: tempFile,
		Username:    "*",
	}).(*HandlerCrond)
	handler.fs = afero.NewMemMapFs()

	expectedJobs := make([]Config, 0, len(testCases))
	for _, testCase := range testCases {
		expectedJobs = append(expectedJobs, testCase.job)

		err := handler.CreateJob(&testCase.job, testCase.schedules, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedJobs, scheduled)

	// now delete all schedules
	for _, testCase := range testCases {
		err := handler.RemoveJob(&testCase.job, PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err = handler.Scheduled("")
	require.NoError(t, err)
	assert.Empty(t, scheduled)
}

func TestCreateReadDeleteCrondAfterLogin(t *testing.T) {
	hourly := calendar.NewEvent(func(e *calendar.Event) {
		e.Minute.MustAddValue(0)
		e.Second.MustAddValue(0)
	})

	tempFile := filepath.Join(t.TempDir(), "crontab")
	handler := NewHandler(SchedulerCrond{
		CrontabFile: tempFile,
		Username:    "*",
	}).(*HandlerCrond)
	handler.fs = afero.NewMemMapFs()

	// after-login only (no calendar schedule)
	loginOnly := Config{
		ProfileName:      "self",
		CommandName:      "backup",
		Command:          "/bin/resticprofile",
		Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "run-schedule", "backup@self"}),
		WorkingDirectory: "/resticprofile",
		ConfigFile:       "examples/dev.yaml",
		Permission:       constants.SchedulePermissionUserLoggedOn,
		AfterLogin:       true,
	}
	require.NoError(t, handler.CreateJob(&loginOnly, nil, PermissionUserLoggedOn))

	// after-login combined with a calendar schedule
	both := Config{
		ProfileName:      "other",
		CommandName:      "check",
		Command:          "/bin/resticprofile",
		Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "run-schedule", "check@other"}),
		WorkingDirectory: "/resticprofile",
		ConfigFile:       "examples/dev.yaml",
		Schedules:        []string{"*-*-* *:00:00"},
		Permission:       constants.SchedulePermissionUserLoggedOn,
		AfterLogin:       true,
	}
	require.NoError(t, handler.CreateJob(&both, []*calendar.Event{hourly}, PermissionUserLoggedOn))

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)
	require.Len(t, scheduled, 2)

	byName := make(map[string]Config, len(scheduled))
	for _, cfg := range scheduled {
		byName[cfg.ProfileName] = cfg
	}

	require.Contains(t, byName, "self")
	assert.True(t, byName["self"].AfterLogin)
	assert.Empty(t, byName["self"].Schedules)

	require.Contains(t, byName, "other")
	assert.True(t, byName["other"].AfterLogin)
	assert.Equal(t, []string{"*-*-* *:00:00"}, byName["other"].Schedules)

	// after-login requires user_logged_on
	require.ErrorContains(t,
		handler.CreateJob(&loginOnly, nil, PermissionSystem),
		"after-login")
}

func TestDetectPermissionCrond(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"", "user", false},
		{"something", "user", false},
		{"system", "system", true},
		{"user", "user", true},
		{"user_logged_on", "user_logged_on", true},
		{"user_logged_in", "user_logged_on", true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(SchedulerCrond{}).(*HandlerCrond)
			perm, safe := handler.DetectSchedulePermission(PermissionFromConfig(fixture.input))
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}

func TestCheckPermission(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		permission Permission
		euid       int
		expected   bool
	}{
		{
			name:       "PermissionUserLoggedOn",
			permission: PermissionUserLoggedOn,
			euid:       1000, // non-root user
			expected:   true,
		},
		{
			name:       "PermissionUserBackground",
			permission: PermissionUserBackground,
			euid:       1000, // non-root user
			expected:   true,
		},
		{
			name:       "PermissionSystem as root",
			permission: PermissionSystem,
			euid:       0, // root user
			expected:   true,
		},
		{
			name:       "PermissionSystem as non-root",
			permission: PermissionSystem,
			euid:       1000, // non-root user
			expected:   false,
		},
		{
			name:       "Undefined permission as root",
			permission: PermissionFromConfig("undefined"),
			euid:       0, // root user
			expected:   true,
		},
		{
			name:       "Undefined permission as non-root",
			permission: PermissionFromConfig("undefined"),
			euid:       1000, // non-root user
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			user := user.User{
				Uid: tc.euid,
			}

			handler := NewHandler(SchedulerCrond{}).(*HandlerCrond)
			result := handler.CheckPermission(user, tc.permission)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNeedsUserInCronEntry(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		config       SchedulerCrond
		job          Config
		expectedUser string
	}{
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "user",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "system",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "-",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "user",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "-",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "system",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "*",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "user",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "*",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "system",
			},
			expectedUser: "",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "testuser",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "user",
			},
			expectedUser: "testuser",
		},
		{
			config: SchedulerCrond{
				CrontabFile: "somefile",
				Username:    "testuser",
			},
			job: Config{
				ProfileName:      "self",
				CommandName:      "check",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "profiles.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Schedules:        []string{"*-*-* *:00:00"},
				ConfigFile:       "profiles.yaml",
				Permission:       "system",
			},
			expectedUser: "testuser",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("config:%s permission:%s expected:%s", tc.config.Username, tc.job.Permission, tc.expectedUser), func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			handler := NewHandler(tc.config).(*HandlerCrond)
			handler.fs = fs
			err := handler.CreateJob(&tc.job, []*calendar.Event{calendar.NewEvent(func(e *calendar.Event) {
				e.Minute.MustAddValue(0)
				e.Second.MustAddValue(0)
			})}, PermissionFromConfig(tc.job.Permission))
			require.NoError(t, err)

			crontab := crond.NewCrontab(nil).
				SetFile(tc.config.CrontabFile).
				SetBinary(tc.config.CrontabBinary).
				SetFs(fs)
			entries, err := crontab.GetEntries()
			require.NoError(t, err)
			require.Len(t, entries, 1)

			if tc.config.Username == "*" && tc.expectedUser == "" {
				assert.NotEmpty(t, entries[0].User())
			} else {
				assert.Equal(t, tc.expectedUser, entries[0].User())
			}
		})
	}
}
