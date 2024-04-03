package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixUnixPaths(t *testing.T) {
	if platform.IsWindows() {
		t.SkipNow()
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	usr, err := user.Current()
	require.NoError(t, err)

	paths := []struct {
		source   string
		expected string
	}{
		{"", ""},
		{"dir", "/prefix/dir"},
		{"/dir", "/dir"},
		{"~/dir", filepath.Join(home, "dir")},
		{"~" + usr.Username + "/dir", filepath.Join(home, "dir")},
		{"~" + usr.Username, home},
		{"~", home},
		{"~file", "/prefix/~file"},
		{"$TEMP_TEST_DIR/dir", "/home/dir"},
		{"some file.txt", "/prefix/some file.txt"},
	}

	err = os.Setenv("TEMP_TEST_DIR", "/home")
	require.NoError(t, err)

	for _, testPath := range paths {
		fixed := fixPath(testPath.source, expandEnv, absolutePrefix("/prefix"), expandUserHome)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
		// running it again should not change the value
		fixed = fixPath(fixed, expandEnv, absolutePrefix("/prefix"))
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
	}
}

func TestFixWindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	paths := []struct {
		source   string
		expected string
	}{
		{``, ``},
		{`dir`, `c:\prefix\dir`},
		{`\dir`, `c:\prefix\dir`},
		{`c:\dir`, `c:\dir`},
		{`~\dir`, filepath.Join(home, "dir")},
		{`~/dir`, home + `/dir`},
		{`~`, home},
		{`~file`, `c:\prefix\~file`},
		{`%TEMP_TEST_DIR%\dir`, `%TEMP_TEST_DIR%\dir`},
		{`${TEMP_TEST_DIR}\dir`, `c:\home\dir`},
		{"some file.txt", `c:\prefix\some file.txt`},
	}

	err = os.Setenv("TEMP_TEST_DIR", "c:\\home")
	require.NoError(t, err)

	for _, testPath := range paths {
		fixed := fixPath(testPath.source, expandEnv, absolutePrefix("c:\\prefix"), expandUserHome)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
		// running it again should not change the value
		fixed = fixPath(fixed, expandEnv, absolutePrefix("c:\\prefix"))
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
	}
}

func TestExpandEnv(t *testing.T) {
	path := os.Getenv("PATH")
	assert.Equal(t, path, expandEnv("$PATH"))
	assert.Equal(t, path, expandEnv("${PATH}"))
	assert.Equal(t, "%PATH%", expandEnv("%PATH%"))
	assert.Equal(t, "$PATH", expandEnv("$$PATH"))
	assert.Equal(t, "", expandEnv("${__UNDEFINED_ENV_VAR__}"))
	assert.Equal(t, "", expandEnv("$__UNDEFINED_ENV_VAR__"))
}

func TestEvaluateSymlinks(t *testing.T) {
	if platform.IsWindows() {
		t.SkipNow()
	}

	var rawDir, dir string
	setup := func(t *testing.T) {
		var err error
		rawDir = t.TempDir()
		dir, err = filepath.EvalSymlinks(rawDir)
		require.NoError(t, err)
	}

	link := func(t *testing.T, path, linkname string) {
		_ = os.Mkdir(filepath.Join(rawDir, path), 0700)
		require.NoError(t, os.Symlink(filepath.Join(rawDir, path), filepath.Join(rawDir, linkname)))
	}

	t.Run("existing-target", func(t *testing.T) {
		setup(t)
		link(t, "a", "b")
		assert.Equal(t, filepath.Join(dir, "a"), evaluateSymlinks(filepath.Join(rawDir, "b")))
		assert.Equal(t, filepath.Join(dir, "a"), evaluateSymlinks(filepath.Join(rawDir, "a")))
		assert.Equal(t, filepath.Join(dir, "a"), evaluateSymlinks(filepath.Join(dir, "a")))
	})

	t.Run("non-existing-target", func(t *testing.T) {
		setup(t)
		link(t, "a", "b")
		assert.Equal(t, filepath.Join(dir, "a", "missing"), evaluateSymlinks(filepath.Join(rawDir, "b", "missing")))
		assert.Equal(t, filepath.Join(dir, "missing/path"), evaluateSymlinks(filepath.Join(rawDir, "missing/path")))
	})

	t.Run("non-existing-targets", func(t *testing.T) {
		setup(t)
		link(t, "a", "b")
		assert.Equal(t, filepath.Join(dir, "a/mis/s/ing"), evaluateSymlinks(filepath.Join(rawDir, "b/mis/s/ing")))
	})

	t.Run("nested", func(t *testing.T) {
		setup(t)
		link(t, "a", "b")
		link(t, "d", "c")
		link(t, "a/nested", "b/c")
		link(t, "d", "b/c/toD")
		assert.Equal(t, filepath.Join(dir, "a/nested"), evaluateSymlinks(filepath.Join(rawDir, "b/c")))
		assert.Equal(t, filepath.Join(dir, "d"), evaluateSymlinks(filepath.Join(rawDir, "b/c/toD")))
		assert.Equal(t, filepath.Join(dir, "d"), evaluateSymlinks(filepath.Join(rawDir, "a/nested/toD")))
	})

	t.Run("usage-in-profile", func(t *testing.T) {
		setup(t)
		link(t, "my-base", "linked-base")
		baseDir := filepath.Join(rawDir, "linked-base")

		config := func(relative bool) string {
			return fmt.Sprintf(`
				[profile]
				base-dir = %q
				[profile.backup]
				source-relative = %v
				source-base = %q
			`, baseDir, relative, baseDir)
		}

		profile, err := getResolvedProfile("toml", config(false), "profile")
		require.NoError(t, err)
		assert.Equal(t, baseDir, profile.BaseDir)
		assert.Equal(t, profile.BaseDir, profile.Backup.SourceBase)

		profile, err = getResolvedProfile("toml", config(true), "profile")
		require.NoError(t, err)
		assert.Equal(t, evaluateSymlinks(baseDir), profile.BaseDir)
		assert.Equal(t, profile.BaseDir, profile.Backup.SourceBase)
	})
}
