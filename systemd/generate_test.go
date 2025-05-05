//go:build !darwin && !windows

package systemd

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/user"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testStandardUser = user.User{Uid: 1001, Gid: 1001, Username: "testuser", UserHomeDir: "/home/testuser", SudoHomeDir: "/home/testuser"}
	testSudoUser     = user.User{Uid: 1001, Gid: 1001, Username: "testuser", UserHomeDir: "/home/testuser", Sudo: true, SudoHomeDir: "/root"}
	testRootUser     = user.User{Uid: 0, Gid: 0, Username: "root", UserHomeDir: "/root", SudoHomeDir: "/root"}
)

func TestGenerateSystemUnit(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)

	err := Unit{fs: fs}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)
}

func TestGenerateSystemUnitServiceAfterNetworkOnline(t *testing.T) {
	const expectedService = `[Unit]
Description=job description
After=network-online.target

[Service]
Type=notify
WorkingDirectory=workdir
ExecStart=commandLine
User=testuser
Environment="HOME=%s"
`
	t.Parallel()
	fs := afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:        "commandLine",
		WorkingDirectory:   "workdir",
		Title:              "name",
		SubTitle:           "backup",
		JobDescription:     "job description",
		TimerDescription:   "timer description",
		Schedules:          []string{"daily"},
		UnitType:           SystemUnit,
		Priority:           "low",
		AfterNetworkOnline: true,
		User:               testStandardUser.Username,
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)

	service, err := afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedService, testSudoUser.UserHomeDir), string(service))
}

func TestGenerateUserUnit(t *testing.T) {
	const expectedService = `[Unit]
Description=job description

[Service]
Type=notify
WorkingDirectory=workdir
ExecStart=commandLine
Environment="HOME=%s"
`
	const expectedTimer = `[Unit]
Description=timer description

[Timer]
OnCalendar=daily
Unit=resticprofile-backup@profile-name.service
Persistent=true

[Install]
WantedBy=timers.target
`
	t.Parallel()
	fs := afero.NewMemMapFs()

	systemdUserDir := filepath.Join(testStandardUser.UserHomeDir, ".config", "systemd", "user")
	serviceFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)

	err := Unit{fs: fs, user: testStandardUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         UserUnit,
		Priority:         "low",
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)

	service, err := afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedService, testStandardUser.UserHomeDir), string(service))

	timer, err := afero.ReadFile(fs, timerFile)
	require.NoError(t, err)
	assert.Equal(t, expectedTimer, string(timer))
}

func TestGenerateUnitTemplateNotFound(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		UnitFile:         "unit-file",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateNotFound(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		TimerFile:        "timer-file",
	})
	require.Error(t, err)
}

func TestGenerateUnitTemplateFailed(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		UnitFile:         "unit",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateFailed(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		TimerFile:        "timer",
	})
	require.Error(t, err)
}

func TestGenerateUnitTemplateFailedToExecute(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		UnitFile:         "unit",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateFailedToExecute(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		TimerFile:        "timer",
	})
	require.Error(t, err)
}

func TestGenerateFromUserDefinedTemplates(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)

	err := afero.WriteFile(fs, "unit", []byte("{{ .JobDescription }}"), 0600)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "timer", []byte("{{ .TimerDescription }}"), 0600)
	require.NoError(t, err)

	err = Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		UnitFile:         "unit",
		TimerFile:        "timer",
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)
}

func TestGenerateWithDropInFile(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	dropInFileContents := []byte(`
[Service]
Environment=foo=bar
`)
	dropInTimerFileContents := []byte(`
[Timer]
RandomizedDelaySec=5h
`)

	require.NoError(t, afero.WriteFile(fs, "98-example.conf", dropInTimerFileContents, 0o600))
	require.NoError(t, afero.WriteFile(fs, "99-example.conf", dropInFileContents, 0o600))

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")
	dropInFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/99-example.resticprofile.conf")
	dropInTimerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer.d/98-example.resticprofile.conf")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)
	assertNoFileExists(t, fs, dropInFile)
	assertNoFileExists(t, fs, dropInTimerFile)

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		DropInFiles:      []string{"98-example.conf", "99-example.conf"},
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)
	requireFileExists(t, fs, dropInFile)
	requireFileExists(t, fs, dropInTimerFile)

	dropIn, err := afero.ReadFile(fs, dropInFile)
	require.NoError(t, err)
	assert.Equal(t, dropInFileContents, dropIn)

	dropIn, err = afero.ReadFile(fs, dropInTimerFile)
	require.NoError(t, err)
	assert.Equal(t, dropInTimerFileContents, dropIn)
}

func TestGenerateCleansUpOrphanDropIns(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()

	dropInFileContents := []byte(`
[Service]
Environment=foo=bar
`)
	dropInTimerFileContents := []byte(`
[Timer]
RandomizedDelaySec=5h
`)

	require.NoError(t, afero.WriteFile(fs, "98-example.conf", dropInTimerFileContents, 0o600))
	require.NoError(t, afero.WriteFile(fs, "99-example.conf", dropInFileContents, 0o600))

	systemdDir := GetSystemDir()

	orphanFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/88-orphan.resticprofile.conf")
	require.NoError(t, afero.WriteFile(fs, orphanFile, []byte{}, 0o600))
	orphanTimerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer.d/87-orphan.resticprofile.conf")
	require.NoError(t, afero.WriteFile(fs, orphanTimerFile, []byte{}, 0o600))

	externallyCreatedFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/77-external.conf")
	require.NoError(t, afero.WriteFile(fs, externallyCreatedFile, []byte{}, 0o600))
	externallyCreatedTimerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer.d/76-external.conf")
	require.NoError(t, afero.WriteFile(fs, externallyCreatedTimerFile, []byte{}, 0o600))

	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")
	dropInFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/99-example.resticprofile.conf")
	dropInTimerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer.d/98-example.resticprofile.conf")

	assertNoFileExists(t, fs, serviceFile)
	assertNoFileExists(t, fs, timerFile)
	assertNoFileExists(t, fs, dropInFile)
	assertNoFileExists(t, fs, dropInTimerFile)

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		DropInFiles:      []string{"98-example.conf", "99-example.conf"},
	})
	require.NoError(t, err)
	requireFileExists(t, fs, serviceFile)
	requireFileExists(t, fs, timerFile)
	requireFileExists(t, fs, dropInFile)
	requireFileExists(t, fs, dropInTimerFile)
	assertNoFileExists(t, fs, orphanFile)
	assertNoFileExists(t, fs, orphanTimerFile)
	requireFileExists(t, fs, externallyCreatedFile)
	requireFileExists(t, fs, externallyCreatedTimerFile)
}

func TestGetUserDirOnReadOnlyFs(t *testing.T) {
	t.Parallel()
	fs := afero.NewReadOnlyFs(afero.NewMemMapFs())
	_, err := Unit{fs: fs, user: testStandardUser}.GetUserDir()
	assert.Error(t, err)
}

func TestGenerateOnReadOnlyFs(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	unit := Unit{fs: fs, user: testSudoUser}
	_, err := unit.GetUserDir()
	assert.NoError(t, err)
	// now make the FS readonly
	fs = afero.NewReadOnlyFs(fs)
	unit = Unit{fs: fs, user: testSudoUser}

	err = unit.Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
	})
	require.Error(t, err)
}

func TestGeneratePriorityFields(t *testing.T) {
	t.Parallel()
	jobName := "name"
	jobCommand := "backup"
	testCases := []struct {
		config      Config
		contains    []string
		notContains []string
	}{
		{
			config: Config{
				CommandLine:      "commandLine",
				WorkingDirectory: "workdir",
				Title:            jobName,
				SubTitle:         jobCommand,
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "",
			},
			contains:    []string{"ExecStart=commandLine\n"},
			notContains: []string{"CPUSchedulingPolicy=", "Nice=", "IOSchedulingClass=", "IOSchedulingPriority="},
		},
		{
			config: Config{
				CommandLine:      "commandLine",
				WorkingDirectory: "workdir",
				Title:            jobName,
				SubTitle:         jobCommand,
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "standard",
			},
			contains:    []string{"ExecStart=commandLine\n"},
			notContains: []string{"CPUSchedulingPolicy=", "Nice=", "IOSchedulingClass=", "IOSchedulingPriority="},
		},
		{
			config: Config{
				CommandLine:      "commandLine",
				WorkingDirectory: "workdir",
				Title:            jobName,
				SubTitle:         jobCommand,
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "background",
			},
			contains:    []string{"CPUSchedulingPolicy=idle\n"},
			notContains: []string{"Nice=", "IOSchedulingClass=", "IOSchedulingPriority="},
		},
		{
			config: Config{
				CommandLine:      "commandLine",
				WorkingDirectory: "workdir",
				Title:            jobName,
				SubTitle:         jobCommand,
				Schedules:        []string{"daily"},
				UnitType:         SystemUnit,
				Priority:         "standard",
				Nice:             10,
			},
			contains:    []string{"Nice=10\n"},
			notContains: []string{"CPUSchedulingPolicy=", "IOSchedulingClass=", "IOSchedulingPriority="},
		},
		{
			config: Config{
				CommandLine:          "commandLine",
				WorkingDirectory:     "workdir",
				Title:                jobName,
				SubTitle:             jobCommand,
				Schedules:            []string{"daily"},
				UnitType:             SystemUnit,
				IOSchedulingClass:    2,
				IOSchedulingPriority: 7,
			},
			contains:    []string{"IOSchedulingClass=2\n", "IOSchedulingPriority=7\n"},
			notContains: []string{"CPUSchedulingPolicy=", "Nice="},
		},
		{
			config: Config{
				CommandLine:          "commandLine",
				WorkingDirectory:     "workdir",
				Title:                jobName,
				SubTitle:             jobCommand,
				Schedules:            []string{"daily"},
				UnitType:             SystemUnit,
				Priority:             "background",
				Nice:                 11,
				IOSchedulingClass:    3,
				IOSchedulingPriority: 7,
			},
			contains:    []string{"CPUSchedulingPolicy=idle\n", "Nice=11\n", "IOSchedulingClass=3\n", "IOSchedulingPriority=7\n"},
			notContains: []string{},
		},
	}

	systemdDir := GetSystemDir()
	baseFile := fmt.Sprintf("resticprofile-%s@profile-%s", jobCommand, jobName)
	serviceFile := filepath.Join(systemdDir, baseFile+".service")

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()

			assertNoFileExists(t, fs, serviceFile)

			err := Unit{fs: fs, user: testSudoUser}.Generate(testCase.config)
			require.NoError(t, err)

			requireFileExists(t, fs, serviceFile)

			contents, err := afero.ReadFile(fs, serviceFile)
			require.NoError(t, err)

			if len(testCase.contains) > 0 {
				for _, c := range testCase.contains {
					assert.Contains(t, string(contents), c)
				}
			}
			if len(testCase.notContains) > 0 {
				for _, c := range testCase.notContains {
					assert.NotContains(t, string(contents), c)
				}
			}
		})
	}
}

func TestGenerateUserField(t *testing.T) {
	t.Parallel()
	fs := afero.NewMemMapFs()
	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")

	err := Unit{fs: fs, user: testSudoUser}.Generate(Config{
		JobDescription:   "Test",
		CommandLine:      "resticprofile",
		WorkingDirectory: "/tmp",
		Title:            "name",
		SubTitle:         "backup",
		UnitType:         SystemUnit,
		User:             "user",
	})
	require.NoError(t, err)

	contents, err := afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)

	expected := `[Unit]
Description=Test

[Service]
Type=notify
WorkingDirectory=/tmp
ExecStart=resticprofile
User=user
Environment="HOME=%s"
`
	assert.Equal(t, fmt.Sprintf(expected, testSudoUser.UserHomeDir), string(contents))
}

func assertNoFileExists(t *testing.T, fs afero.Fs, filename string) {
	t.Helper()
	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.Falsef(t, exists, "file %q exists", filename)
}

func requireFileExists(t *testing.T, fs afero.Fs, filename string) {
	t.Helper()
	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	require.Truef(t, exists, "file %q does not exist", filename)
}
