//go:build darwin

package schedule

import (
	"bytes"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
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
				Permission:       constants.SchedulePermissionUserLoggedOn,
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

func TestParsePrint(t *testing.T) {
	const launchctlOutput = `user/503/local.resticprofile.self.check = {
	active count = 0
	path = /Users/cp/Library/LaunchAgents/local.resticprofile.self.check.agent.plist
	type = LaunchAgent
	state = not running

	program = /Users/cp/go/src/github.com/creativeprojects/resticprofile/resticprofile
	arguments = {
		/Users/cp/go/src/github.com/creativeprojects/resticprofile/resticprofile
		--no-prio
		--no-ansi
		--config
		examples/dev.yaml
		run-schedule
		check@self
	}

	working directory = /Users/cp/go/src/github.com/creativeprojects/resticprofile

	stdout path = local.resticprofile.self.check.log
	stderr path = local.resticprofile.self.check.log
	default environment = {
		PATH => /usr/bin:/bin:/usr/sbin:/sbin
	}

	environment = {
		PATH => /usr/local/bin:/System/Cryptexes/App/usr/bin:/usr/bin:/bin:/usr/sbin:/sbin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/local/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/bin:/var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/appleinternal/bin:/Users/cp/go/bin
		RESTICPROFILE_SCHEDULE_ID => examples/dev.yaml:check@self
		XPC_SERVICE_NAME => local.resticprofile.self.check
	}

	domain = user/503
	asid = 100008
	minimum runtime = 10
	exit timeout = 5
	nice = 0
	runs = 1
	last exit code = 0

	event triggers = {
		local.resticprofile.self.check.267436470 => {
			keepalive = 0
			service = local.resticprofile.self.check
			stream = com.apple.launchd.calendarinterval
			monitor = com.apple.UserEventAgent-Aqua
			descriptor = {
				"Minute" => 5
			}
		}
	}

	event channels = {
		"com.apple.launchd.calendarinterval" = {
			port = 0xc1851
			active = 0
			managed = 1
			reset = 0
			hide = 0
			watching = 1
		}
	}

	spawn type = daemon (3)
	jetsam priority = 40
	jetsam memory limit (active) = (unlimited)
	jetsam memory limit (inactive) = (unlimited)
	jetsamproperties category = daemon
	jetsam thread limit = 32
	cpumon = default
	probabilistic guard malloc policy = {
		activation rate = 1/1000
		sample rate = 1/0
	}

	properties = 
}
`

	t.Parallel()

	info := parsePrintStatus([]byte(launchctlOutput))
	assertMapHasKeys(t, info, launchctlPrintKeys)

}

func assertMapHasKeys(t *testing.T, source map[string]string, keys []string) {
	t.Helper()

	for _, key := range keys {
		if _, found := source[key]; !found {
			t.Errorf("key %q not found in map, available keys are: %s", key, strings.Join(slices.Collect(maps.Keys(source)), ", "))
		}
	}
}

func TestIsServiceRegistered(t *testing.T) {
	services := []struct {
		domain       string
		name         string
		isRegistered bool
	}{
		{"system", "service.that.surely.does.not.exist", false},
		{"system", "com.apple.fseventsd", true},
	}

	for _, service := range services {
		registered, err := isServiceRegistered(service.domain, service.name)
		require.NoError(t, err)
		assert.Equal(t, service.isRegistered, registered)
	}
}

func TestParsePrintSystemService(t *testing.T) {
	// this test should tell us when the output format of the launchctl print command is changing
	cmd := launchctlCommand(launchdPrint, "system/com.apple.fseventsd")
	output, err := cmd.Output()
	require.NoError(t, err)

	info := parsePrintStatus(output)
	assert.Greater(t, len(info), 20) // keep a low number to avoid flaky test
	assert.Equal(t, "system", info["domain"])
}

func TestDetectPermissionLaunchd(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		input    string
		expected string
		safe     bool
	}{
		{"", "user_logged_on", true},
		{"something", "user_logged_on", true},
		{"system", "system", true},
		{"user", "user", true},
		{"user_logged_on", "user_logged_on", true},
		{"user_logged_in", "user_logged_on", true}, // I did the typo as I was writing the doc, so let's add it here :)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.input, func(t *testing.T) {
			t.Parallel()

			handler := NewHandler(SchedulerLaunchd{}).(*HandlerLaunchd)
			perm, safe := handler.DetectSchedulePermission(PermissionFromConfig(fixture.input))
			assert.Equal(t, fixture.expected, perm.String())
			assert.Equal(t, fixture.safe, safe)
		})
	}
}
