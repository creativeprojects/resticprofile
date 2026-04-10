//go:build !windows

package util

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	helperExecutableCommand = "executable"
	testTimeout             = 5 * time.Second
)

func TestExecutable(t *testing.T) {
	helpersPath := os.Getenv("TEST_HELPERS")
	if helpersPath == "" {
		helpersPath = "../build"
	}
	helperBinary := filepath.Join(helpersPath, platform.Executable("test-args"))
	helperBinary, err := filepath.Abs(helperBinary)
	require.NoError(t, err, "Failed to get absolute path of helper binary", helperBinary)
	require.FileExists(t, helperBinary, "Helper binary is not available at expected path")

	tempDir := t.TempDir()
	symlinkBinary := filepath.Join(tempDir, "executable_test_symlink")
	err = os.Symlink(helperBinary, symlinkBinary)
	require.NoError(t, err, "Failed to create symlink for helper binary")

	t.Run("absolute", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, helperBinary, helperExecutableCommand)
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+helperBinary+"\"\n", string(output), "Output should match the helper binary path")
	})

	t.Run("absolute symlink", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, symlinkBinary, helperExecutableCommand)
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+symlinkBinary+"\"\n", string(output), "Output should match the helper binary path")
	})

	t.Run("relative", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./"+filepath.Base(helperBinary), helperExecutableCommand)
		cmd.Dir = filepath.Dir(helperBinary) // Set the working directory to the helper binary's directory
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+helperBinary+"\"\n", string(output), "Output should match the helper binary path")
	})

	t.Run("relative symlink", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./"+filepath.Base(symlinkBinary), helperExecutableCommand)
		cmd.Dir = tempDir // Set the working directory to the temp directory
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+symlinkBinary+"\"\n", string(output), "Output should match the helper binary path")
	})

	t.Run("from PATH", func(t *testing.T) {
		path := os.Getenv("PATH")
		t.Setenv("PATH", filepath.Dir(helperBinary)+string(os.PathListSeparator)+path) // Add tempDir to PATH for this test
		t.Logf("Using PATH: %s", os.Getenv("PATH"))

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, filepath.Base(helperBinary), helperExecutableCommand)
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+helperBinary+"\"\n", string(output), "Output should match the helper binary path")
	})

	t.Run("symlink from PATH", func(t *testing.T) {
		path := os.Getenv("PATH")
		t.Setenv("PATH", tempDir+string(os.PathListSeparator)+path) // Add tempDir to PATH for this test
		t.Logf("Using PATH: %s", os.Getenv("PATH"))

		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		cmd := exec.CommandContext(ctx, filepath.Base(symlinkBinary), helperExecutableCommand)
		output, err := cmd.Output()
		require.NoError(t, err)

		t.Log(string(output))
		assert.Equal(t, "\""+symlinkBinary+"\"\n", string(output), "Output should match the helper binary path")
	})
}
