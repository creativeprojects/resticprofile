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
	"syscall"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	helperBinary string
)

func TestMain(m *testing.M) {
	// using an anonymous function to handle defer statements before os.Exit()
	exitCode := func() int {
		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultTestTimeout)
		defer cancel()

		tempDir, err := os.MkdirTemp("", "resticprofile-lock")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot create temp dir: %v\n", err)
			return 1
		}
		fmt.Printf("using temporary dir: %q\n", tempDir)
		defer os.RemoveAll(tempDir)

		helperBinary = filepath.Join(tempDir, platform.Executable("locktest"))

		cmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", helperBinary, "./test")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error building helper binary: %s\n", err)
			return 1
		}

		return m.Run()
	}()
	os.Exit(exitCode)
}

func getTempfile(t *testing.T) string {
	t.Helper()

	tempfile := filepath.Join(t.TempDir(), fmt.Sprintf("%s.tmp", t.Name()))
	t.Log("Using temporary file", tempfile)
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

func TestProcessPID(t *testing.T) {
	t.Parallel()

	var childPID int32
	buffer := &bytes.Buffer{}

	// use the lock helper binary (we only need to wait for some time, we don't need the locking part)
	cmd := shell.NewCommand(helperBinary, []string{"-wait", "200", "-lock", filepath.Join(t.TempDir(), t.Name())})
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
	cmd := shell.NewCommand(helperBinary, []string{"-wait", "100", "-lock", filepath.Join(t.TempDir(), t.Name())})
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

	if platform.IsWindows() {
		t.Skip("cannot send a signal to a child process in Windows")
	}
	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	cmd := exec.Command(helperBinary, "-wait", "1", "-lock", lockfile)
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Run()
	assert.NoError(t, err)
	assert.Equal(t, "lock acquired\ntask finished\nlock released\n", buffer.String())
}

func TestLockIsRemovedAfterInterruptSignal(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("cannot send a signal to a child process in Windows")
	}
	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	cmd := exec.Command(helperBinary, "-wait", "2000", "-lock", lockfile)
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	require.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
	err = cmd.Process.Signal(syscall.SIGINT)
	require.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)
	assert.Equal(t, "lock acquired\ntask interrupted\nlock released\n", buffer.String())
}

func TestLockIsRemovedAfterInterruptSignalInsideShell(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("cannot send a signal to a child process in Windows")
	}
	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	cmd := exec.Command("sh", "-c", "exec "+helperBinary+" -wait 2000 -lock "+lockfile)
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	require.NoError(t, err)

	time.Sleep(300 * time.Millisecond)
	err = cmd.Process.Signal(syscall.SIGINT)
	require.NoError(t, err)

	err = cmd.Wait()
	assert.NoError(t, err)
	assert.Equal(t, "lock acquired\ntask interrupted\nlock released\n", buffer.String())
}
