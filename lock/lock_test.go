package lock

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/stretchr/testify/assert"
)

func TestLockIsAvailable(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestLockIsAvailable", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
}

func TestLockIsNotAvailable(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestLockIsNotAvailable", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
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
	assert.Regexp(t, regexp.MustCompile(`^[\-\\\w]+ on \w+, \d+-\w+-\d+ \d+:\d+:\d+ \w* from [\.\-\w]+$`), who)
}

func TestNoPID(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestNoPID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
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
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestSetPID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
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
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestSetMorePID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
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

// This test is using the shell package. This is just a convenient wrapper around cmd.exe and sh
func TestProcessFinished(t *testing.T) {
	childPID := 0
	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := shell.NewSignalledCommand("echo", []string{"Hello World!"}, c)
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int) {
		childPID = pid
	}
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// at that point, the child process should be finished
	running, err := process.PidExists(int32(childPID))
	assert.NoError(t, err)
	assert.False(t, running)
}

// This test is using the shell package. This is just a convenient wrapper around cmd.exe and sh
func TestProcessNotFinished(t *testing.T) {
	childPID := 0
	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	// use ping to make sure the process is running for long enough to check its existence
	var parameters []string
	if runtime.GOOS == "windows" {
		// it will run for 1 second
		parameters = []string{"-n", "2", "127.0.0.1"}
	} else {
		// run for 200ms (don't need a whole second)
		// 0.2 is the minimum in linux, 0.1 in darwin
		parameters = []string{"-c", "2", "-i", "0.2", "127.0.0.1"}
	}

	cmd := shell.NewSignalledCommand("ping", parameters, c)
	cmd.Stdout = buffer
	// SetPID method is called right after we forked and have a PID available
	cmd.SetPID = func(pid int) {
		childPID = pid
		running, err := process.PidExists(int32(childPID))
		assert.NoError(t, err)
		assert.True(t, running)
	}
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	// at that point, the child process should be finished
	running, err := process.PidExists(int32(childPID))
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestForceLockIsAvailable(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestForceLockIsAvailable", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.ForceAcquire())
}

func TestForceLockWithNoPID(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestForceLockWithNoPID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	other := NewLock(tempfile)
	defer other.Release()
	assert.False(t, other.ForceAcquire())
	assert.False(t, other.HasLocked())
}

// This test is using the shell package. This is just a convenient wrapper around cmd.exe and sh
func TestForceLockWithExpiredPID(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestForceLockWithExpiredPID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
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
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	// child process should be finished
	// let's close the lockfile handle manually (unix doesn't actually care, but windows would complain)
	lock.file.Close()

	other := NewLock(tempfile)
	defer other.Release()
	assert.True(t, other.ForceAcquire())
	assert.True(t, other.HasLocked())
}

// This test is using the shell package. This is just a convenient wrapper around cmd.exe and sh
func TestForceLockWithRunningPID(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestForceLockWithRunningPID", time.Now().UnixNano(), os.Getpid()))
	t.Log("Using temporary file", tempfile)
	lock := NewLock(tempfile)
	defer lock.Release()

	assert.True(t, lock.TryAcquire())
	assert.True(t, lock.HasLocked())

	// run a child process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	// use ping to make sure the process is running for long enough to check its existence
	var parameters []string
	if runtime.GOOS == "windows" {
		// it will run for 1 second
		parameters = []string{"-n", "2", "127.0.0.1"}
	} else {
		// run for 200ms (don't need a whole second)
		// 0.2 is the minimum in linux, 0.1 in darwin
		parameters = []string{"-c", "2", "-i", "0.2", "127.0.0.1"}
	}

	cmd := shell.NewSignalledCommand("ping", parameters, c)
	cmd.SetPID = func(pid int) {
		lock.SetPID(pid)
		// make sure we cannot break the lock right now
		other := NewLock(tempfile)
		defer other.Release()
		assert.False(t, other.ForceAcquire())
		assert.False(t, other.HasLocked())
	}
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
