package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/lock"
	"github.com/stretchr/testify/assert"
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
