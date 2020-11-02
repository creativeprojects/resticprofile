package lock

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// SetPID is a callback that writes the PID in the lockfile
type SetPID func(pid int)

// Lock prevents code to run at the same time by using a lockfile
type Lock struct {
	Lockfile string
	file     *os.File
	locked   bool
}

// NewLock creates a new lock
func NewLock(filename string) *Lock {
	return &Lock{
		Lockfile: filename,
		locked:   false,
	}
}

// TryAcquire returns true if the lock was successfully set. It returns false if a lock already exists
func (l *Lock) TryAcquire() bool {
	return l.lock()
}

// ForceAcquire returns true if the lock was successfully set.
//
// If a lock file already exists, it reads the PID of the supposidly running process and check if
// it's still running. If there no more process with this PID, it will try to delete the lock, then
// recreate a new lock.
func (l *Lock) ForceAcquire() bool {
	// try nicely first
	if l.lock() {
		return true
	}
	pid, err := l.LastPID()
	if err != nil {
		return false
	}
	running, err := process.PidExists(pid)
	if err != nil || running {
		return false
	}
	// it looks like the process is no longer running, try to delete the lockfile
	err = os.Remove(l.Lockfile)
	if err != nil {
		return false
	}
	// previous lock file removed, let's make a new one now
	return l.lock()
}

// Release the lockfile
func (l *Lock) Release() {
	if l.file != nil {
		_ = l.file.Close()
	}
	l.unlock()
}

// Who owns the lock?
func (l *Lock) Who() (string, error) {
	buffer, err := ioutil.ReadFile(l.Lockfile)
	if err != nil {
		return "", err
	}
	// first line should be "who" owns the lock, any subsequent line will contain the restic PIDs
	contents := strings.Split(string(buffer), "\n")
	return contents[0], nil
}

// SetPID writes down the PID in the lock file.
// You can run the method as many times as you want when the PID changes
func (l *Lock) SetPID(pid int) {
	if !l.locked {
		return
	}
	// just add the PID on a newline
	_, _ = l.file.WriteString(fmt.Sprintf("\n%d", pid))
}

// HasLocked check this instance (and only this one) has locked the file
func (l *Lock) HasLocked() bool {
	return l.locked
}

// LastPID returns the last PID written into the lock file.
func (l *Lock) LastPID() (int32, error) {
	buffer, err := ioutil.ReadFile(l.Lockfile)
	if err != nil {
		return 0, err
	}
	// first line should be "who" owns the lock, any subsequent line will contain the restic PIDs
	contents := strings.Split(string(buffer), "\n")
	// we stop at line 1: line 0 should not contain any PID
	for i := len(contents) - 1; i >= 1; i-- {
		if contents[i] != "" {
			pid, err := strconv.ParseInt(contents[i], 10, 32)
			if err == nil {
				return int32(pid), nil
			}
		}
	}
	return 0, errors.New("lock file does not contain any child process information")
}

func (l *Lock) lock() bool {
	var err error

	l.file, err = os.OpenFile(l.Lockfile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return false
	}
	// Leave the lock file open

	l.locked = true

	username := "unknown user"
	currentUser, err := user.Current()
	if err == nil {
		username = currentUser.Username
	}

	hostname := "unknown hostname"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}

	now := time.Now().Format(time.RFC850)

	// No error checking... it's not a big deal if we cannot write that
	_, _ = l.file.WriteString(fmt.Sprintf("%s on %s from %s", username, now, hostname))
	return true
}

func (l *Lock) unlock() {
	_ = os.Remove(l.Lockfile)
	l.locked = false
}
