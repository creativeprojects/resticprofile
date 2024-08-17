package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockBinary string
	echoBinary string
)

func TestMain(m *testing.M) {
	mockBinary = "./shellmock"
	if platform.IsWindows() {
		mockBinary = `.\\shellmock.exe`
	}
	cmdMock := exec.Command("go", "build", "-buildvcs=false", "-o", mockBinary, "./shell/mock")
	if output, err := cmdMock.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error building shell/mock binary: %s\nCommand output: %s\n", err, string(output))
	}

	echoBinary = "./shellecho"
	if platform.IsWindows() {
		echoBinary = `.\\shellecho.exe`
	}
	cmdEcho := exec.Command("go", "build", "-buildvcs=false", "-o", echoBinary, "./shell/echo")
	if output, err := cmdEcho.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error building shell/echo binary: %s\nCommand output: %s\n", err, string(output))
	}

	exitCode := m.Run()
	_ = os.Remove(mockBinary)
	_ = os.Remove(echoBinary)
	os.Exit(exitCode)
}

func TestGetProfile(t *testing.T) {
	baseDir, _ := filepath.Abs(t.TempDir())
	baseDir, err := filepath.EvalSymlinks(baseDir)
	require.NoError(t, err)
	baseDir = filepath.ToSlash(baseDir)

	c, err := config.Load(strings.NewReader(`
		[default]
		 repository = "test-repo"
		 tag = ["{{ .CurrentDir }}", "{{ .StartupDir }}"]
		[with-base]
		 inherit = "default"
		 base-dir = "`+baseDir+`"
		[with-invalid-base]
		 inherit = "default"
		 base-dir = "~/some-dir-not-exists"
	`), "toml")
	require.NoError(t, err)

	getWd := func(t *testing.T) string {
		dir, err := os.Getwd()
		require.NoError(t, err)
		return filepath.ToSlash(dir)
	}

	cwd := getWd(t)

	getProf := func(t *testing.T, name string) (profile *config.Profile, cleanup func()) {
		var err error
		profile, cleanup, err = openProfile(c, name)
		require.NoError(t, err)
		require.NotNil(t, cleanup)
		require.NotNil(t, profile)
		return
	}

	t.Run("default", func(t *testing.T) {
		profile, cleanup := getProf(t, "default")
		assert.Equal(t, "test-repo", profile.Repository.Value())
		assert.Equal(t, []any{cwd, cwd}, profile.OtherFlags[constants.ParameterTag])

		assert.Equal(t, cwd, getWd(t))
		cleanup()
		assert.Equal(t, cwd, getWd(t))
	})

	t.Run("with-base-dir", func(t *testing.T) {
		profile, cleanup := getProf(t, "with-base")
		assert.Equal(t, "test-repo", profile.Repository.Value())
		assert.Equal(t, []any{baseDir, cwd}, profile.OtherFlags[constants.ParameterTag])

		assert.Equal(t, baseDir, getWd(t))
		cleanup()
		assert.Equal(t, cwd, getWd(t))
	})

	t.Run("load-error", func(t *testing.T) {
		profile, cleanup, err := openProfile(c, "unknown")
		require.NotNil(t, cleanup)
		defer cleanup()

		assert.Nil(t, profile)
		assert.EqualError(t, err, "cannot load profile 'unknown': not found")
	})

	t.Run("with-base-dir-error", func(t *testing.T) {
		profile, cleanup, err := openProfile(c, "with-invalid-base")
		require.NotNil(t, cleanup)
		defer cleanup()

		require.NotNil(t, profile)
		dir := filepath.FromSlash(profile.BaseDir)
		assert.ErrorContains(t, err, `cannot change to base directory "`+dir+`" in profile "with-invalid-base": chdir `+dir+`: `)
	})
}
