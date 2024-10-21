package schedule

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
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
	if platform.IsWindows() {
		t.Skip("crond scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerCrond{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}

func TestSystemdInit(t *testing.T) {
	if platform.IsWindows() || platform.IsDarwin() {
		t.Skip("systemd scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerSystemd{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
func TestLaunchdInit(t *testing.T) {
	if !platform.IsDarwin() {
		t.Skip("launchd scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerLaunchd{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
func TestWindowsInit(t *testing.T) {
	if !platform.IsWindows() {
		t.Skip("windows scheduler is not supported on this platform")
	}
	scheduler := NewScheduler(NewHandler(SchedulerWindows{}), "profile")
	err := scheduler.Init()
	defer scheduler.Close()
	require.NoError(t, err)
}
