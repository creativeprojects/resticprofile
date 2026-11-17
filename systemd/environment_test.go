//go:build !darwin && !windows && !openbsd && !netbsd && !freebsd

package systemd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testServiceFile     = "resticprofile-backup@profile-name.service"
	testEnvironmentFile = "resticprofile-backup@profile-name.service.d/resticprofile.env"
)

func testEnvironmentConfig(environment ...string) Config {
	return Config{
		CommandLine:      "commandLine",
		WorkingDirectory: "workdir",
		Title:            "name",
		SubTitle:         "backup",
		Schedules:        []string{"daily"},
		UnitType:         SystemUnit,
		Environment:      environment,
	}
}

func TestGenerateEnvironmentFile(t *testing.T) {
	t.Setenv("SUDO_USER", "")

	fs := afero.NewMemMapFs()
	unit := Unit{fs: fs, user: testSudoUser}
	serviceFile := filepath.Join(GetSystemDir(), testServiceFile)
	environmentFile := filepath.Join(GetSystemDir(), testEnvironmentFile)

	// a file from a previous version is replaced
	require.NoError(t, afero.WriteFile(fs, environmentFile, []byte("old"), 0o644))

	config := testEnvironmentConfig(
		"RESTICPROFILE_SCHEDULE_ID=profiles.yaml:backup@name",
		"RESTIC_PASSWORD=pa$$ \"w`o\\rd'%h\nnext line",
		"HOME=/captured",
	)
	require.NoError(t, unit.Generate(config))

	service, err := afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)
	assert.Contains(t, string(service), "Environment=\"HOME=/root\"\nEnvironmentFile="+environmentFile+"\n")
	assert.NotContains(t, string(service), "RESTIC")

	info, err := fs.Stat(environmentFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	content, err := afero.ReadFile(fs, environmentFile)
	require.NoError(t, err)
	assert.Equal(t, "RESTICPROFILE_SCHEDULE_ID=\"profiles.yaml:backup@name\"\n"+
		"RESTIC_PASSWORD=\"pa\\$\\$ \\\"w\\`o\\\\rd'%h\nnext line\"\n", string(content))

	// the generated HOME keeps precedence over the captured one
	read, err := unit.Read(testServiceFile, SystemUnit)
	require.NoError(t, err)
	assert.Equal(t, []string{config.Environment[0], config.Environment[1], "HOME=/root"}, read.Environment)

	// nothing captured: no file
	require.NoError(t, unit.Generate(testEnvironmentConfig()))
	service, err = afero.ReadFile(fs, serviceFile)
	require.NoError(t, err)
	assert.NotContains(t, string(service), "EnvironmentFile")
	_, err = fs.Stat(environmentFile)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestEnvironmentFileDoesNotFollowSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	require.NoError(t, os.WriteFile(target, []byte("unchanged"), 0o644))

	link := filepath.Join(root, GetSystemDir(), testEnvironmentFile)
	require.NoError(t, os.MkdirAll(filepath.Dir(link), 0o755))
	require.NoError(t, os.Symlink(target, link))

	unit := Unit{fs: afero.NewBasePathFs(afero.NewOsFs(), root), user: testSudoUser}
	require.NoError(t, unit.Generate(testEnvironmentConfig("RESTIC_PASSWORD=secret")))

	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "unchanged", string(content))

	info, err := os.Lstat(link)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestParseEnvironmentFile(t *testing.T) {
	t.Parallel()

	environment := []string{"A=", "B=plain", "C=\\ \" ` $ ' %h \n end"}
	content := ""
	for _, env := range environment {
		name, value, _ := strings.Cut(env, "=")
		content += name + "=" + quoteEnvironmentValue(value) + "\n"
	}
	assert.Equal(t, environment, parseEnvironmentFile(content))
}
