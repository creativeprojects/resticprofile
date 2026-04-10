//go:build linux

package priority

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartProcessWithIOPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	helpersPath := os.Getenv("TEST_HELPERS")
	if helpersPath == "" {
		helpersPath = "../build"
	}
	testBinary := filepath.Join(helpersPath, platform.Executable("test-args"))
	testBinary, err := filepath.Abs(testBinary)
	require.NoError(t, err, "Failed to get absolute path of test-args helper binary", testBinary)
	require.FileExists(t, testBinary, "test-args helper binary is not available at expected path")

	// Run these 3 tests inside one test, so we don't have concurrency issue
	t.Run("WithNormalIOPriority", func(t *testing.T) {

		output, err := runChildProcess(testBinary)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		assert.Contains(t, output, "IOPriority: class = 0, value = 4")

	})

	t.Run("WithBestEffortIOPriority", func(t *testing.T) {

		err := SetIONice(2, 6)
		if err != nil {
			t.Fatal(err)
		}

		output, err := runChildProcess(testBinary)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		assert.Contains(t, output, "IOPriority: class = 2, value = 6")

	})

	t.Run("WithIdlePriority", func(t *testing.T) {

		err := SetIONice(3, 4)
		if err != nil {
			t.Fatal(err)
		}

		output, err := runChildProcess(testBinary)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(output)
		assert.Contains(t, output, "IOPriority: class = 3, value = 4")

	})
}
