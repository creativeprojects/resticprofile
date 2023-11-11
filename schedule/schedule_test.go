package schedule

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

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
	scheduler := NewScheduler(NewHandler(SchedulerDefaultOS{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}

func TestCrondInit(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("crond scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerCrond{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}

func TestSystemdInit(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("systemd scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerSystemd{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
func TestLaunchdInit(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("launchd scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerLaunchd{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
func TestWindowsInit(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerWindows{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
