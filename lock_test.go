package main

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/lock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockRunWithNoLockfile(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	err := lockRun("", false, nil, nil, callback)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestLockRunWithNoLock(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	assert.NoFileExists(t, lockfile)

	err := lockRun(lockfile, false, nil, nil, callback)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
	assert.NoFileExists(t, lockfile)
}

func TestLockRunWithLock(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	err := os.WriteFile(lockfile, []byte{}, 0o600)
	assert.NoError(t, err)
	assert.FileExists(t, lockfile)

	err = lockRun(lockfile, false, nil, nil, callback)
	assert.Error(t, err)
	assert.Equal(t, 0, called)
	assert.FileExists(t, lockfile)
}

func TestLockRunWithLockAndForce(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	err := os.WriteFile(lockfile, []byte{}, 0o600)
	assert.NoError(t, err)
	assert.FileExists(t, lockfile)

	err = lockRun(lockfile, true, nil, nil, callback)
	assert.Error(t, err)
	assert.Equal(t, 0, called)
	assert.FileExists(t, lockfile)
}

func TestLockRunWithLockAndWait(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	err := os.WriteFile(lockfile, []byte{}, 0o600)
	assert.NoError(t, err)
	assert.FileExists(t, lockfile)

	// remove the lock after half a second
	timer := time.AfterFunc(500*time.Millisecond, func() {
		err := os.Remove(lockfile)
		assert.NoError(t, err)
	})
	defer timer.Stop()

	wait := 1 * time.Second
	err = lockRun(lockfile, false, &wait, nil, callback)
	assert.NoError(t, err)
	assert.Equal(t, 1, called)
}

func TestLockRunWithLockAndCancel(t *testing.T) {
	called := 0
	callback := func(setPID lock.SetPID) error {
		called++
		return nil
	}
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	err := os.WriteFile(lockfile, []byte{}, 0o600)
	assert.NoError(t, err)
	assert.FileExists(t, lockfile)

	sigChan := make(chan os.Signal, 1)
	// cancel the wait after half a second
	timer := time.AfterFunc(500*time.Millisecond, func() {
		sigChan <- os.Interrupt
	})
	defer timer.Stop()

	wait := 1 * time.Second
	err = lockRun(lockfile, false, &wait, sigChan, callback)
	assert.Error(t, err)
	assert.Equal(t, 0, called)
	assert.FileExists(t, lockfile)
}

func TestLockRunReleasesLockImmediatelyOnSignal(t *testing.T) {
	// The lock must disappear while run() is still executing — not only after
	// lockRun returns via defer. That is the behaviour that prevents lockfiles
	// from surviving a follow-up SIGKILL during cleanup.
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	sigChan := make(chan os.Signal, 1) // non-nil enables the termination watcher

	inject := make(chan os.Signal, 1)
	orig := openTermChan
	openTermChan = func() (<-chan os.Signal, func()) {
		return inject, func() {}
	}
	t.Cleanup(func() { openTermChan = orig })

	releasedDuringRun := false
	err := lockRun(lockfile, false, nil, sigChan, func(setPID lock.SetPID) error {
		require.FileExists(t, lockfile)
		inject <- syscall.SIGTERM

		deadline := time.After(2 * time.Second)
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-deadline:
				return errors.New("lockfile was not released during run() after SIGTERM")
			case <-ticker.C:
				if _, err := os.Stat(lockfile); os.IsNotExist(err) {
					releasedDuringRun = true
					return nil
				}
			}
		}
	})
	require.NoError(t, err)
	assert.True(t, releasedDuringRun)
	assert.NoFileExists(t, lockfile)
}

func TestLockRunReleasesLockOnNormalReturn(t *testing.T) {
	lockfile := filepath.Join(t.TempDir(), "lockfile")
	sigChan := make(chan os.Signal, 1)

	sawLock := false
	err := lockRun(lockfile, false, nil, sigChan, func(setPID lock.SetPID) error {
		if _, err := os.Stat(lockfile); err == nil {
			sawLock = true
		}
		return nil
	})
	require.NoError(t, err)
	assert.True(t, sawLock)
	assert.NoFileExists(t, lockfile)
}
