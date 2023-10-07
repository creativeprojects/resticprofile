//go:build !darwin && !windows

package systemd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSystemUnit(t *testing.T) {
	fs = afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)

	err := Generate(Config{
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
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
}

func TestGenerateSystemUnitTimerAfterNetworkOnline(t *testing.T) {
	const expectedTimer = `[Unit]
Description=timer description
After=network-online.target

[Timer]
OnCalendar=daily
Unit=resticprofile-backup@profile-name.service
Persistent=true

[Install]
WantedBy=timers.target
`

	fs = afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)

	err := Generate(Config{
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
	})
	require.NoError(t, err)
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)

	timer, err := afero.ReadFile(fs, timerFile)
	require.NoError(t, err)
	assert.Equal(t, expectedTimer, string(timer))
}

func TestGenerateUserUnit(t *testing.T) {
	const expectedService = `[Unit]
Description=job description

[Service]
Type=notify
WorkingDirectory=workdir
ExecStart=commandLine
Nice=5
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
	fs = afero.NewMemMapFs()
	u, err := user.Current()
	require.NoError(t, err)
	systemdUserDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	serviceFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.timer")
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)

	err = Generate(Config{
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
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)

	service, err := afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedService, home), string(service))

	timer, err := afero.ReadFile(fs, timerFile)
	require.NoError(t, err)
	assert.Equal(t, expectedTimer, string(timer))
}

func TestGenerateUnitTemplateNotFound(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := Generate(Config{
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
	fs = afero.NewMemMapFs()

	err := Generate(Config{
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
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Generate(Config{
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
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Generate(Config{
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
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Generate(Config{
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
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Generate(Config{
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
	fs = afero.NewMemMapFs()

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)

	err := afero.WriteFile(fs, "unit", []byte("{{ .JobDescription }}"), 0600)
	require.NoError(t, err)
	err = afero.WriteFile(fs, "timer", []byte("{{ .TimerDescription }}"), 0600)
	require.NoError(t, err)

	err = Generate(Config{
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
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
}

func TestGenerateWithDropInFile(t *testing.T) {
	fs = afero.NewMemMapFs()

	dropInFileContents := []byte(`
[Service]
Environment=foo=bar
`)

	err := afero.WriteFile(fs, "99-example.conf", dropInFileContents, 0o600)
	require.NoError(t, err)

	systemdDir := GetSystemDir()
	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")
	dropInFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/99-example.resticprofile.conf")

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)
	assertNoFileExists(t, dropInFile)

	err = Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		DropInFiles:      []string{"99-example.conf"},
	})
	require.NoError(t, err)
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
	requireFileExists(t, dropInFile)

	dropIn, err := afero.ReadFile(fs, dropInFile)
	require.NoError(t, err)
	assert.Equal(t, dropInFileContents, dropIn)
}

func TestGenerateCleansUpOrphanDropIns(t *testing.T) {
	fs = afero.NewMemMapFs()

	dropInFileContents := []byte(`
[Service]
Environment=foo=bar
`)

	err := afero.WriteFile(fs, "99-example.conf", dropInFileContents, 0o600)
	require.NoError(t, err)

	systemdDir := GetSystemDir()

	orphanFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/88-orphan.resticprofile.conf")
	err = afero.WriteFile(fs, orphanFile, []byte{}, 0o600)
	require.NoError(t, err)

	externallyCreatedFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/77-external.conf")
	err = afero.WriteFile(fs, externallyCreatedFile, []byte{}, 0o600)
	require.NoError(t, err)

	serviceFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.timer")
	dropInFile := filepath.Join(systemdDir, "resticprofile-backup@profile-name.service.d/99-example.resticprofile.conf")

	assertNoFileExists(t, serviceFile)
	assertNoFileExists(t, timerFile)
	assertNoFileExists(t, dropInFile)

	err = Generate(Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		JobDescription:   "job description",
		TimerDescription: "timer description",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Priority:         "low",
		DropInFiles:      []string{"99-example.conf"},
	})
	require.NoError(t, err)
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
	requireFileExists(t, dropInFile)
	assertNoFileExists(t, orphanFile)
	requireFileExists(t, externallyCreatedFile)

	dropIn, err := afero.ReadFile(fs, dropInFile)
	require.NoError(t, err)
	assert.Equal(t, dropInFileContents, dropIn)
}

func TestGetUserDirOnReadOnlyFs(t *testing.T) {
	fs = afero.NewReadOnlyFs(afero.NewMemMapFs())
	_, err := GetUserDir()
	assert.Error(t, err)
}

func TestGenerateOnReadOnlyFs(t *testing.T) {
	fs = afero.NewMemMapFs()
	_, err := GetUserDir()
	assert.NoError(t, err)
	// now make the FS readonly
	fs = afero.NewReadOnlyFs(fs)

	err = Generate(Config{
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

func assertNoFileExists(t *testing.T, filename string) {
	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	assert.Falsef(t, exists, "file %q exists", filename)
}

func requireFileExists(t *testing.T, filename string) {
	exists, err := afero.Exists(fs, filename)
	require.NoError(t, err)
	require.Truef(t, exists, "file %q does not exist", filename)
}
