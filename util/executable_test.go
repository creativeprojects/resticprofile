package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutableIsAbsolute(t *testing.T) {
	executable, err := Executable()
	require.NoError(t, err)
	assert.NotEmpty(t, executable)

	assert.True(t, filepath.IsAbs(executable))
}

func TestExecutable(t *testing.T) {
	if platform.IsWindows() {
		t.Skip("Executable test is not applicable on Windows")
	}

	tempDir, err := os.MkdirTemp("", "resticprofile-executable")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	})

	helperBinary := filepath.Join(tempDir, "executable_test_helper")
	assert.True(t, filepath.IsAbs(helperBinary), "Helper binary path should be absolute")

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", helperBinary, "./test_executable")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Error building helper binary: %s\n", err)
	}

	symlinkBinary := filepath.Join(tempDir, "executable_test_symlink")
	err = os.Symlink(helperBinary, symlinkBinary)
	require.NoError(t, err, "Failed to create symlink for helper binary")

	t.Run("absolute", func(t *testing.T) {
		cmd = exec.Command(helperBinary)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+helperBinary+"\"\n", "Output should match the helper binary path")
	})

	t.Run("absolute symlink", func(t *testing.T) {
		cmd = exec.Command(symlinkBinary)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+symlinkBinary+"\"\n", "Output should match the helper binary path")
	})

	t.Run("relative", func(t *testing.T) {
		cmd = exec.Command("./" + filepath.Base(helperBinary))
		cmd.Dir = tempDir // Set the working directory to the temp directory
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+helperBinary+"\"\n", "Output should match the helper binary path")
	})

	t.Run("relative symlink", func(t *testing.T) {
		cmd = exec.Command("./" + filepath.Base(symlinkBinary))
		cmd.Dir = tempDir // Set the working directory to the temp directory
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+symlinkBinary+"\"\n", "Output should match the helper binary path")
	})

	t.Run("from PATH", func(t *testing.T) {
		path := os.Getenv("PATH")
		t.Cleanup(func() {
			os.Setenv("PATH", path) // Restore original PATH after test
		})
		os.Setenv("PATH", tempDir+string(os.PathListSeparator)+path) // Add tempDir to PATH for this test
		t.Logf("Using PATH: %s", os.Getenv("PATH"))

		cmd = exec.Command(filepath.Base(helperBinary))
		cmd.Dir = tempDir // Set the working directory to the temp directory
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+helperBinary+"\"\n", "Output should match the helper binary path")
	})

	t.Run("symlink from PATH", func(t *testing.T) {
		path := os.Getenv("PATH")
		t.Cleanup(func() {
			os.Setenv("PATH", path) // Restore original PATH after test
		})
		os.Setenv("PATH", tempDir+string(os.PathListSeparator)+path) // Add tempDir to PATH for this test
		t.Logf("Using PATH: %s", os.Getenv("PATH"))

		cmd = exec.Command(filepath.Base(symlinkBinary))
		cmd.Dir = tempDir // Set the working directory to the temp directory
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Error executing helper binary: %s\n", err)
		}
		t.Log(string(output))
		assert.Equal(t, string(output), "\""+symlinkBinary+"\"\n", "Output should match the helper binary path")
	})
}
