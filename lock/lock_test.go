package lock

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	lockBinary string
)

func TestMain(m *testing.M) {
	exitCode := func() int {
		var err error
		helpersPath := os.Getenv("TEST_HELPERS")
		if helpersPath == "" {
			helpersPath = "../build"
		}
		lockBinary = filepath.Join(helpersPath, platform.Executable("test-args"))
		lockBinary, err = filepath.Abs(lockBinary)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get absolute path of test-args binary: %v\n", err)
			return 1
		}
		if _, err := os.Stat(lockBinary); err != nil {
			fmt.Fprintf(os.Stderr, "test-args binary is not available at expected path: %s\n", lockBinary)
			return 1
		}

		return m.Run()
	}()
	os.Exit(exitCode)
}

func getTempfile(t *testing.T) string {
	t.Helper()

	tempfile := filepath.Join(t.TempDir(), fmt.Sprintf("%s.tmp", t.Name()))
	return tempfile
}

func TestLockIsAvailable(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
}

func TestLockIsNotAvailable(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	other := NewLock(tempfile)
	defer other.Release()
	assert.False(t, other.TryAcquire())
	assert.False(t, other.HasLocked())

	who, err := other.Who()
	assert.NoError(t, err)
	assert.NotEmpty(t, who)
	assert.Regexp(t, regexp.MustCompile(`^[\.\-\\\w]+ on \w+, \d+-\w+-\d+ \d+:\d+:\d+ \w* from [\.\-\w]+$`), who)
}

func TestNoPID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()
	lock.TryAcquire()

	other := NewLock(tempfile)
	defer other.Release()

	pid, err := other.LastPID()
	assert.Error(t, err)
	assert.Empty(t, pid)
}

func TestSetOnePID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()
	lock.TryAcquire()
	lock.SetPID(11)

	other := NewLock(tempfile)
	defer other.Release()

	pid, err := other.LastPID()
	assert.NoError(t, err)
	assert.Equal(t, int32(11), pid)
}

func TestSetMorePID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()
	lock.TryAcquire()
	lock.SetPID(11)
	lock.SetPID(12)
	lock.SetPID(13)

	other := NewLock(tempfile)
	defer other.Release()

	pid, err := other.LastPID()
	assert.NoError(t, err)
	assert.Equal(t, int32(13), pid)
}

func TestForceLockIsAvailable(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.ForceAcquire())
}

func TestForceLockWithNoPID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	other := NewLock(tempfile)
	defer other.Release()
	assert.False(t, other.ForceAcquire())
	assert.False(t, other.HasLocked())
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

func TestForceLockWithRunningPID(t *testing.T) {
	t.Parallel()

	tempfile := getTempfile(t)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	// user the lock helper binary (we only need to wait for some time, we don't need the locking part)
	cmd := shell.NewCommand(lockBinary, []string{"lock", "-wait", "100", "-lock", filepath.Join(t.TempDir(), t.Name())})
	cmd.SetPID = func(pid int32) {
		lock.SetPID(pid)
		// make sure we cannot break the lock right now
		other := NewLock(tempfile)
		defer other.Release()
		assert.False(t, other.ForceAcquire())
		assert.False(t, other.HasLocked())
	}
	_, _, err := cmd.Run()
	require.NoError(t, err)
}

func TestLockWithNoInterruption(t *testing.T) {
	t.Parallel()

	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, lockBinary, "lock", "-wait", "10", "-lock", lockfile)
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Run()
	assert.NoError(t, err)
	assert.Equal(t, "started\nlock acquired\ntask finished\nlock released\n", buffer.String())
}
