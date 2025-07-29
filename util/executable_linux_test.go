//go:build linux

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveExecutable(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		path, err := resolveExecutable("")
		assert.Equal(t, "", path)
		assert.Error(t, err)
		assert.Equal(t, "executable path is empty", err.Error())
	})

	t.Run("absolute path", func(t *testing.T) {
		path, err := resolveExecutable("/usr/bin/ls")
		assert.NoError(t, err)
		assert.Equal(t, "/usr/bin/ls", path)
	})

	t.Run("relative path with dot", func(t *testing.T) {
		wd, err := os.Getwd()
		require.NoError(t, err)

		path, err := resolveExecutable("./test")
		assert.NoError(t, err)
		expected := filepath.Join(wd, "test")
		assert.Equal(t, expected, path)
	})

	t.Run("command in PATH", func(t *testing.T) {
		// Testing with "ls" which should be available on most Linux systems
		path, err := resolveExecutable("ls")
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
		t.Log(path)
		// The exact path can vary by system, but it should be an absolute path
		assert.True(t, filepath.IsAbs(path), "Path should be absolute")
	})

	t.Run("command not in PATH", func(t *testing.T) {
		path, err := resolveExecutable("this_command_should_not_exist_anywhere")
		assert.Equal(t, "", path)
		assert.Error(t, err)
	})
}
