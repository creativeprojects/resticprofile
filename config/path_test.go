package config

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixUnixPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	paths := []struct {
		source   string
		expected string
	}{
		{"", ""},
		{"dir", "/prefix/dir"},
		{"/dir", "/dir"},
		{"~/dir", "~/dir"},
		{"$TEMP_TEST_DIR/dir", "/home/dir"},
		{"some file.txt", "/prefix/some\\ file.txt"},
		{"/**/.git", "/\\*\\*/.git"},
		{"/\\*\\*/.git", "/\\*\\*/.git"},
		{`/?`, `/\?`},
		{`/\?`, `/\?`},
		{`/\\?`, `/\\\?`},
		{`/\\\?`, `/\\\?`},
		{`/\\\\?`, `/\\\\\?`},
		{`/ ?*`, `/\ \?\*`},
	}

	err := os.Setenv("TEMP_TEST_DIR", "/home")
	require.NoError(t, err)

	for _, testPath := range paths {
		fixed := fixPath(testPath.source, expandEnv, absolutePrefix("/prefix"), escapeShellString)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
		// running it again should not change the value
		fixed = fixPath(fixed, expandEnv, absolutePrefix("/prefix"), escapeShellString)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
	}
}

func TestFixWindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}

	paths := []struct {
		source   string
		expected string
	}{
		{``, ``},
		{`dir`, `\prefix\dir`},
		{`\dir`, `\prefix\dir`},
		{`c:\dir`, `c:\dir`},
		{`%TEMP_TEST_DIR%\dir`, `%TEMP_TEST_DIR%\dir`},
		{"some file.txt", `\prefix\some file.txt`},
	}

	err := os.Setenv("TEMP_TEST_DIR", "/home")
	require.NoError(t, err)

	for _, testPath := range paths {
		fixed := fixPath(testPath.source, expandEnv, absolutePrefix("\\prefix"), escapeShellString)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
		// running it again should not change the value
		fixed = fixPath(fixed, expandEnv, absolutePrefix("\\prefix"), escapeShellString)
		assert.Equalf(t, testPath.expected, fixed, "source was '%s'", testPath.source)
	}
}
