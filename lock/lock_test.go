package lock

import (
	"fmt"
	"os"
	"path/filepath"
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

	other := NewLock(tempfile)
	defer other.Release()
	assert.False(t, other.TryAcquire())

	who := other.Who()
	t.Log("File was locked by", other.Who())
	assert.NotEmpty(t, who)
}
