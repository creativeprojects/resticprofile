package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/lock"
)

// openTermChan returns a channel that receives process termination signals.
// Tests replace this to inject signals without sending real OS signals (which
// would terminate the go test process).
var openTermChan = func() (<-chan os.Signal, func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	return ch, func() { signal.Stop(ch) }
}

// lockRun is making sure the function is only run once by putting a lockfile on the disk
func lockRun(lockFile string, force bool, lockWait *time.Duration, sigChan <-chan os.Signal, run func(setPID lock.SetPID) error) error {
	// No lock
	if lockFile == "" {
		return run(nil)
	}

	// Make sure the path to the lock exists
	if dir := filepath.Dir(lockFile); dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			clog.Warningf("the profile will run without a lockfile: %v", err)
			return run(nil)
		}
	}

	// Acquire lock
	runLock := lock.NewLock(lockFile)
	success := runLock.TryAcquire()
	start := time.Now()
	locker := ""
	lockWaitLogged := time.Unix(0, 0)

	for !success {
		if who, err := runLock.Who(); err == nil {
			if locker != who {
				lockWaitLogged = time.Unix(0, 0)
			}
			locker = who
		} else if errors.Is(err, fs.ErrNotExist) {
			locker = "none"
		} else {
			return fmt.Errorf("another process left the lockfile unreadable: %w", err)
		}

		// should we try to force our way?
		if force {
			success = runLock.ForceAcquire()

			if lockWait == nil || success {
				clog.Warningf("previous run of the profile started by %s hasn't finished properly", locker)
			}
		} else {
			success = runLock.TryAcquire()
		}

		// Retry or return?
		if !success {
			if lockWait == nil {
				return fmt.Errorf("another process is already running this profile: %s", locker)
			}
			if time.Since(start) < *lockWait {
				lockName := fmt.Sprintf("%s locked by %s", lockFile, locker)
				lockWaitLogged = logLockWait(lockName, start, lockWaitLogged, 0, *lockWait)

				sleep := min(constants.LocalLockRetryDelay, *lockWait)
				err := interruptibleSleep(sleep, sigChan)
				if err != nil {
					return err
				}
			} else {
				clog.Warningf("previous run of the profile hasn't finished after %s", *lockWait)
				lockWait = nil
			}
		}
	}

	// Run locked.
	//
	// Release the profile lock as soon as a termination signal arrives, not only
	// when run() returns. After SIGINT/SIGTERM the process may still spend time in
	// run-after-fail / finally hooks or waiting on a child; systemd (and similar
	// managers) often follow up with SIGKILL after TimeoutStopSec. Releasing early
	// keeps the lockfile from surviving a hard kill.
	//
	// Use a dedicated signal.Notify channel so we do not compete with the shared
	// sigChan that shell commands use to forward signals to child processes.
	var wg sync.WaitGroup
	stopWatch := make(chan struct{})
	release := func() { runLock.Release() }

	if sigChan != nil {
		termChan, stopTerm := openTermChan()
		defer stopTerm()

		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sig := <-termChan:
				clog.Debugf("received %v while holding profile lock, releasing early", sig)
				release()
			case <-stopWatch:
			}
		}()
	}

	err := run(runLock.SetPID)
	close(stopWatch)
	wg.Wait()
	release()
	return err
}

const logLockWaitEvery = 5 * time.Minute

func logLockWait(lockName string, started, lastLogged time.Time, executed, maxLockWait time.Duration) time.Time {
	now := time.Now()
	lastLog := now.Sub(lastLogged)
	elapsed := now.Sub(started).Truncate(time.Second)
	waited := (elapsed - executed).Truncate(time.Second)
	remaining := (maxLockWait - elapsed).Truncate(time.Second)

	if lastLog > logLockWaitEvery {
		if elapsed > logLockWaitEvery {
			clog.Infof("lock wait (remaining %s / waited %s / elapsed %s): %s", remaining, waited, elapsed, strings.TrimSpace(lockName))
		} else {
			clog.Infof("lock wait (remaining %s): %s", remaining, strings.TrimSpace(lockName))
		}
		return now
	}

	return lastLogged
}
