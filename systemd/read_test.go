//go:build !darwin && !windows

package systemd

import (
	"fmt"
	"path"
	"testing"

	"github.com/creativeprojects/resticprofile/user"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServiceUnit = `[Unit]
Description=resticprofile copy for profile self in examples/linux.yaml
OnFailure=unit-status-mail@%n.service
# Description=commented

[Service]
Type=notify
WorkingDirectory=/home/testuser/go/src/github.com/creativeprojects/resticprofile
ExecStart=/tmp/go-build982790897/b001/exe/resticprofile --no-prio --no-ansi --config examples/linux.yaml run-schedule copy@self
Nice=19
IOSchedulingClass=3
IOSchedulingPriority=7
User=testuser
Environment="RESTICPROFILE_SCHEDULE_ID=examples/linux.yaml:copy@self"
Environment="HOME=/home/testuser"
`
	testTimerUnit = `[Unit]
Description=copy timer for profile self in examples/linux.yaml

[Timer]
OnCalendar=*:45
Unit=resticprofile-copy@profile-self.service
Persistent=true

[Install]
WantedBy=timers.target`
)

func TestReadUnitFile(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	unitFile := "resticprofile-copy@profile-self.service"
	timerFile := "resticprofile-copy@profile-self.timer"
	require.NoError(t, afero.WriteFile(fs, path.Join(systemdSystemDir, unitFile), []byte(testServiceUnit), 0o600))
	require.NoError(t, afero.WriteFile(fs, path.Join(systemdSystemDir, timerFile), []byte(testTimerUnit), 0o600))

	unit := Unit{fs: fs, user: testSudoUser}
	cfg, err := unit.Read(unitFile, SystemUnit)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	expected := &Config{
		CommandLine:          "/tmp/go-build982790897/b001/exe/resticprofile --no-prio --no-ansi --config examples/linux.yaml run-schedule copy@self",
		Environment:          []string{"RESTICPROFILE_SCHEDULE_ID=examples/linux.yaml:copy@self", "HOME=/home/testuser"},
		WorkingDirectory:     testSudoUser.HomeDir + "/go/src/github.com/creativeprojects/resticprofile",
		Title:                "self",
		SubTitle:             "copy",
		JobDescription:       "resticprofile copy for profile self in examples/linux.yaml",
		TimerDescription:     "",
		Schedules:            []string{"*:45"},
		UnitType:             SystemUnit,
		Priority:             "standard",
		UnitFile:             "",
		TimerFile:            "",
		DropInFiles:          []string(nil),
		AfterNetworkOnline:   false,
		Nice:                 19,
		CPUSchedulingPolicy:  "",
		IOSchedulingClass:    3,
		IOSchedulingPriority: 7,
		User:                 testSudoUser.Username,
	}
	assert.Equal(t, expected, cfg)
}

func TestReadSystemUnit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		user   user.User
		config Config
	}{
		{
			user: testRootUser,
			config: Config{
				CommandLine:      "/bin/resticprofile --config profiles.yaml run-schedule backup@profile1",
				WorkingDirectory: "/workdir",
				Title:            "profile1",
				SubTitle:         "backup",
				JobDescription:   "job description",
				TimerDescription: "timer description",
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "background",
			},
		},
		{
			user: testSudoUser,
			config: Config{
				CommandLine:      "/bin/resticprofile --config profiles.yaml run-schedule backup@profile1",
				WorkingDirectory: "/workdir",
				Title:            "profile1",
				SubTitle:         "backup",
				JobDescription:   "job description",
				TimerDescription: "timer description",
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "background",
			},
		},
		{
			user: testStandardUser,
			config: Config{
				CommandLine:      "/bin/resticprofile --no-ansi --config profiles.yaml run-schedule check@profile2",
				WorkingDirectory: "/workdir",
				Title:            "profile2",
				SubTitle:         "check",
				JobDescription:   "",
				TimerDescription: "timer description",
				Schedules:        []string{"daily", "weekly"},
				UnitType:         UserUnit,
				Priority:         "standard",
				Environment: []string{
					"TMP=/tmp",
				},
			},
		},
		{
			user: testSudoUser,
			config: Config{
				CommandLine:      "/bin/resticprofile --no-ansi --config profiles.yaml run-schedule check@profile3",
				WorkingDirectory: "/workdir",
				Title:            "profile3",
				SubTitle:         "forget",
				JobDescription:   "job description",
				TimerDescription: "",
				Schedules:        []string{"monthly"},
				UnitType:         SystemUnit,
				Priority:         "standard",
				User:             testSudoUser.Username,
			},
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			unit := Unit{fs: afero.NewMemMapFs(), user: tc.user}
			baseUnit := fmt.Sprintf("resticprofile-%s@profile-%s", tc.config.SubTitle, tc.config.Title)
			serviceFile := baseUnit + ".service"

			err := unit.Generate(tc.config)
			require.NoError(t, err)

			readCfg, err := unit.Read(serviceFile, tc.config.UnitType)
			require.NoError(t, err)
			assert.NotNil(t, readCfg)

			expected := &Config{
				Title:            tc.config.Title,
				SubTitle:         tc.config.SubTitle,
				JobDescription:   tc.config.JobDescription,
				WorkingDirectory: tc.config.WorkingDirectory,
				CommandLine:      tc.config.CommandLine,
				UnitType:         tc.config.UnitType,
				Environment:      append(tc.config.Environment, "HOME="+tc.user.HomeDir),
				Schedules:        tc.config.Schedules,
				Priority:         tc.config.Priority,
				User:             tc.config.User,
			}
			assert.Equal(t, expected, readCfg)
		})
	}
}
