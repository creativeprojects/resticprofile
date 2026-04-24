package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixWindowsPaths(t *testing.T) {
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
