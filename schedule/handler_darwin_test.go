//go:build darwin

package schedule

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/darwin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

func TestHandlerCrond(t *testing.T) {
	handler := NewHandler(SchedulerCrond{})
	assert.IsType(t, &HandlerCrond{}, handler)
}

func TestPListEncoderWithCalendarInterval(t *testing.T) {
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><array><dict><key>Day</key><integer>1</integer><key>Hour</key><integer>0</integer></dict></array></plist>`
	event := calendar.NewEvent(func(e *calendar.Event) {
		_ = e.Day.AddValue(1)
		_ = e.Hour.AddValue(0)
	})
	entry := darwin.GetCalendarIntervalsFromSchedules([]*calendar.Event{event})
	buffer := &bytes.Buffer{}
	encoder := plist.NewEncoder(buffer)
	err := encoder.Encode(entry)
	require.NoError(t, err)
	assert.Equal(t, expected, buffer.String())
}

func TestParseStatus(t *testing.T) {
	status := `{
	"StandardOutPath" = "local.resticprofile.self.check.log";
	"LimitLoadToSessionType" = "Aqua";
	"StandardErrorPath" = "local.resticprofile.self.check.log";
	"Label" = "local.resticprofile.self.check";
	"OnDemand" = true;
	"LastExitStatus" = 0;
	"Program" = "/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile";
	"ProgramArguments" = (
		"/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile";
		"--no-ansi";
		"--config";
		"examples/dev.yaml";
		"--name";
		"self";
		"check";
	);
};`
	expected := map[string]string{
		"StandardOutPath":        "local.resticprofile.self.check.log",
		"LimitLoadToSessionType": "Aqua",
		"StandardErrorPath":      "local.resticprofile.self.check.log",
		"Label":                  "local.resticprofile.self.check",
		"OnDemand":               "true",
		"LastExitStatus":         "0",
		"Program":                "/Users/go/src/github.com/creativeprojects/resticprofile/resticprofile",
	}

	output := parseStatus(status)
	assert.Equal(t, expected, output)
}

func TestHandlerInstanceDefault(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.NotNil(t, handler)
}

func TestHandlerInstanceLaunchd(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{})
	assert.NotNil(t, handler)
}

func TestLaunchdJobPreservesEnv(t *testing.T) {
	pathEnv := os.Getenv("PATH")
	fixtures := []struct {
		environment []string
		expected    map[string]string
	}{
		{expected: map[string]string{"PATH": pathEnv}},
		{environment: []string{"path=extra-var"}, expected: map[string]string{"PATH": pathEnv, "path": "extra-var"}},
		{environment: []string{"PATH=custom-path"}, expected: map[string]string{"PATH": "custom-path"}},
	}

	for i, fixture := range fixtures {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			handler := NewHandler(SchedulerLaunchd{}).(*HandlerLaunchd)
			cfg := &Config{ProfileName: "t", CommandName: "s", Environment: fixture.environment}
			launchdJob := handler.getLaunchdJob(cfg, []*calendar.Event{})
			assert.Equal(t, fixture.expected, launchdJob.EnvironmentVariables)
		})
	}
}

func TestCreateUserPlist(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{}).(*HandlerLaunchd)
	handler.fs = afero.NewMemMapFs()

	launchdJob := &darwin.LaunchdJob{
		Label: "TestCreateSystemPlist",
	}
	filename, err := handler.createPlistFile(launchdJob, PermissionUserBackground)
	require.NoError(t, err)

	_, err = handler.fs.Stat(filename)
	assert.NoError(t, err)
}

func TestCreateSystemPlist(t *testing.T) {
	handler := NewHandler(SchedulerLaunchd{}).(*HandlerLaunchd)
	handler.fs = afero.NewMemMapFs()

	launchdJob := &darwin.LaunchdJob{
		Label: "TestCreateSystemPlist",
	}
	filename, err := handler.createPlistFile(launchdJob, PermissionSystem)
	require.NoError(t, err)

	_, err = handler.fs.Stat(filename)
	assert.NoError(t, err)
}

func TestReadingLaunchdScheduled(t *testing.T) {
	calendarEvent := calendar.NewEvent(func(e *calendar.Event) {
		_ = e.Second.AddValue(0)
		_ = e.Minute.AddValue(0)
		_ = e.Minute.AddValue(30)
	})
	testCases := []struct {
		job       Config
		schedules []*calendar.Event
	}{
		{
			job: Config{
				ProfileName:      "testscheduled",
				CommandName:      "backup",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Permission:       constants.SchedulePermissionSystem,
				ConfigFile:       "examples/dev.yaml",
				Schedules:        []string{"*-*-* *:00,30:00"},
			},
			schedules: []*calendar.Event{calendarEvent},
		},
		{
			job: Config{
				ProfileName:      "test.scheduled",
				CommandName:      "backup",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "config file.yaml", "--name", "self", "backup"}),
				WorkingDirectory: "/resticprofile",
				Permission:       constants.SchedulePermissionSystem,
				ConfigFile:       "config file.yaml",
				Schedules:        []string{"*-*-* *:00,30:00"},
			},
			schedules: []*calendar.Event{calendarEvent},
		},
		{
			job: Config{
				ProfileName:      "testscheduled",
				CommandName:      "backup",
				Command:          "/bin/resticprofile",
				Arguments:        NewCommandArguments([]string{"--no-ansi", "--config", "examples/dev.yaml", "--name", "self", "check"}),
				WorkingDirectory: "/resticprofile",
				Permission:       constants.SchedulePermissionUser,
				ConfigFile:       "examples/dev.yaml",
				Schedules:        []string{"*-*-* *:00,30:00"},
			},
			schedules: []*calendar.Event{calendarEvent},
		},
	}

	handler := NewHandler(SchedulerLaunchd{}).(*HandlerLaunchd)
	handler.fs = afero.NewMemMapFs()

	expectedJobs := []Config{}
	for _, testCase := range testCases {
		expectedJobs = append(expectedJobs, testCase.job)

		_, err := handler.createPlistFile(handler.getLaunchdJob(&testCase.job, testCase.schedules), PermissionFromConfig(testCase.job.Permission))
		require.NoError(t, err)
	}

	scheduled, err := handler.Scheduled("")
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedJobs, scheduled)
}
