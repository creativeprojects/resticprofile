//+build !darwin,!windows

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
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"",
		"",
	})
	require.NoError(t, err)
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
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
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		UserUnit,
		"low",
		"",
		"",
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
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"unit-file",
		"",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateNotFound(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := Generate(Config{
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"",
		"timer-file",
	})
	require.Error(t, err)
}

func TestGenerateUnitTemplateFailed(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Generate(Config{
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"unit",
		"",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateFailed(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ ."), 0600)
	require.NoError(t, err)

	err = Generate(Config{
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"",
		"timer",
	})
	require.Error(t, err)
}

func TestGenerateUnitTemplateFailedToExecute(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "unit", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Generate(Config{
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"unit",
		"",
	})
	require.Error(t, err)
}

func TestGenerateTimerTemplateFailedToExecute(t *testing.T) {
	fs = afero.NewMemMapFs()

	err := afero.WriteFile(fs, "timer", []byte("{{ .Toto }}"), 0600)
	require.NoError(t, err)

	err = Generate(Config{
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"",
		"timer",
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
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"unit",
		"timer",
	})
	require.NoError(t, err)
	requireFileExists(t, serviceFile)
	requireFileExists(t, timerFile)
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
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		SystemUnit,
		"low",
		"",
		"",
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
