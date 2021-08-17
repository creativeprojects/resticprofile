package config

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixUnixPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
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
