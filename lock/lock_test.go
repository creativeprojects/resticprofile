package lock

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

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
	assert.Equal(t, fmt.Sprintf("%d", 11), pid)
}

func TestSetMorePID(t *testing.T) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", "TestSetPID", time.Now().UnixNano(), os.Getpid()))
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
	assert.Equal(t, fmt.Sprintf("%d", 13), pid)
}
