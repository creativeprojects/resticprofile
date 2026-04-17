//go:build !netbsd

package lock

import (
	"bytes"
	"os"
	"os/signal"
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

	// use the lock helper binary (we only need to wait for some time, we don't need the locking part)
	cmd := shell.NewCommand(lockBinary, []string{"lock", "-wait", "200", "-lock", filepath.Join(t.TempDir(), t.Name())})
	cmd.Stdout = buffer
	// SetPID method is called right after we forked and have a PID available
	cmd.SetPID = func(pid int32) {
		childPID = pid
		running, err := process.PidExists(childPID)
		assert.NoError(t, err)
		assert.True(t, running)
	}
	_, _, err := cmd.Run()
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

	// run a child process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := shell.NewSignalledCommand("echo", []string{"Hello World!"}, c)
	cmd.SetPID = lock.SetPID
	_, _, err := cmd.Run()
	require.NoError(t, err)

	// child process should be finished
	// let's close the lockfile handle manually (unix doesn't actually care, but windows would complain)
	lock.file.Close()

	other := NewLock(tempfile)
	defer other.Release()
	assert.True(t, other.ForceAcquire())
	assert.True(t, other.HasLocked())
}
