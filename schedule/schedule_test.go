package schedule

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentDirIsAbsoluteOnAllPlatforms(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(cwd))
}

func TestFileIsAbsoluteOnAllPlatforms(t *testing.T) {
	dir, err := filepath.Abs("file.txt")
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(dir))
	t.Log(dir)
}

func TestExecutableIsAbsoluteOnAllPlatforms(t *testing.T) {
	binary, err := os.Executable()
	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(binary))
	t.Log(binary)
}

func TestInit(t *testing.T) {
	err := Init()
	defer Close()
	require.NoError(t, err)
}

func TestCrondInit(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("crond not supported on this platform")
	}
	// TODO: the Scheduler package variable can make things tricky to test
	Scheduler = constants.SchedulerCrond
	err := Init()
	require.NoError(t, err)
	Scheduler = ""
}
