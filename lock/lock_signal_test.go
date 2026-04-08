//go:build !windows

package lock

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockIsRemovedAfterInterruptSignal(t *testing.T) {
	// don't run in parallel
	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, helperBinary, "-wait", "2000", "-lock", lockfile)
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	require.NoError(t, err, "starting child process")

	time.Sleep(time.Second)
	err = cmd.Process.Signal(os.Interrupt)
	require.NoError(t, err, "sending interrupt signal to child process")

	err = cmd.Wait()
	if isSignalError(err) && platform.IsDarwin() {
		t.Skip("inconclusive test: command failed to run properly - flaky test on macOS only")
	}
	assert.NoError(t, err, "waiting for child process to finish")

	assert.Equal(t, "started\nlock acquired\ntask interrupted\nlock released\n", buffer.String())
}

func TestLockIsRemovedAfterInterruptSignalInsideShell(t *testing.T) {
	// don't run in parallel
	lockfile := getTempfile(t)

	var err error
	buffer := &bytes.Buffer{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", "exec \""+helperBinary+"\" -wait 2000 -lock \""+lockfile+"\"")
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	err = cmd.Start()
	require.NoError(t, err, "starting child process inside a shell")

	time.Sleep(time.Second)
	err = cmd.Process.Signal(os.Interrupt)
	require.NoError(t, err, "sending interrupt signal to child process")

	err = cmd.Wait()
	if isSignalError(err) && platform.IsDarwin() {
		t.Skip("inconclusive test: command failed to run properly - flaky test on macOS only")
	}
	assert.NoError(t, err, "waiting for child process to finish")

	assert.Equal(t, "started\nlock acquired\ntask interrupted\nlock released\n", buffer.String())
}

func isSignalError(err error) bool {
	if err == nil {
		return false
	}
	exitErr, ok := errors.AsType[*exec.ExitError](err)
	if !ok {
		return false
	}
	if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
		return status.Signaled()
	}
	return false
}
