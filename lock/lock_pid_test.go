//go:build !netbsd

package lock

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessPID(t *testing.T) {
	t.Parallel()

	var childPID int32
	buffer := &bytes.Buffer{}

	// SetPID method is called right after we forked and have a PID available
	setPID := func(pid int) {
		childPID = int32(pid)
		running, err := process.PidExists(childPID)
		assert.NoError(t, err)
		assert.True(t, running)
	}
	// use the lock helper binary (we only need to wait for some time, we don't need the locking part)
	commandRunner := shell.NewDirectRunner(shell.RunnerConfig{})
	err := commandRunner.Run(context.Background(), shell.CommandConfig{
		Command: lockBinary,
		Args:    []string{"lock", "-wait", "200", "-lock", filepath.Join(t.TempDir(), t.Name())},
		Stdout:  buffer,
		SetPID:  setPID,
	})
	require.NoError(t, err)

	// at that point, the child process should be finished
	running, err := process.PidExists(childPID)
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestForceLockWithExpiredPID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	// run a child process to get a real PID
	commandRunner := shell.NewDirectRunner(shell.RunnerConfig{})
	err := commandRunner.Run(context.Background(), shell.CommandConfig{Command: "whoami", SetPID: lock.SetPID})
	require.NoError(t, err)

	// child process should be finished
	// let's close the lockfile handle manually (unix doesn't actually care, but windows would complain)
	lock.file.Close()

	other := NewLock(tempfile)
	defer other.Release()
	assert.True(t, other.ForceAcquire())
	assert.True(t, other.HasLocked())
}
