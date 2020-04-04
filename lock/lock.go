package lock

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"time"
)

// Lock prevents code to run at the same time by using a lockfile
type Lock struct {
	Lockfile string
	file     *os.File
}

// NewLock creates a new lock
func NewLock(filename string) *Lock {
	return &Lock{
		Lockfile: filename,
	}
}

// TryAcquire returns true if the lock was successfully set. It returns false if a lock already exists
func (l *Lock) TryAcquire() bool {
	return l.lock()
}

// Release the lockfile
func (l *Lock) Release() {
	if l.file != nil {
		l.file.Close()
	}
	l.unlock()
}

// Who owns the lock?
func (l *Lock) Who() string {
	who, err := ioutil.ReadFile(l.Lockfile)
	if err != nil {
		return err.Error()
	}
	return string(who)
}

func (l *Lock) lock() bool {
	var err error

	l.file, err = os.OpenFile(l.Lockfile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return false
	}
	// Leave the lock file open

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
	l.file.WriteString(fmt.Sprintf("%s on %s from %s", username, now, hostname))
	return true
}

func (l *Lock) unlock() {
	os.Remove(l.Lockfile)
}
