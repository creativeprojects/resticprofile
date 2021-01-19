//+build !darwin,!windows

package systemd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
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
	u, err := user.Current()
	require.NoError(t, err)
	systemdUserDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	serviceFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.service")
	timerFile := filepath.Join(systemdUserDir, "resticprofile-backup@profile-name.timer")
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	defer func() {
		os.Remove(serviceFile)
		os.Remove(timerFile)
	}()
	assert.NoFileExists(t, serviceFile)
	assert.NoFileExists(t, timerFile)

	err = Generate(
		"commandLine",
		"workdir",
		"name",
		"backup",
		"job description",
		"timer description",
		[]string{"daily"},
		UserUnit,
		"low",
	)
	require.NoError(t, err)
	require.FileExists(t, serviceFile)
	require.FileExists(t, timerFile)

	service, err := ioutil.ReadFile(serviceFile)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(expectedService, home), string(service))

	timer, err := ioutil.ReadFile(timerFile)
	require.NoError(t, err)
	assert.Equal(t, expectedTimer, string(timer))
}
