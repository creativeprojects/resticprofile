package shell

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const ResticLockFailureOutput = `
unable to create lock in backend: repository is already locked by PID 27153 on app-server-01 by root (UID 0, GID 0)
lock was created at 2021-04-17 19:00:16 (1s727.5ms ago)
storage ID 870530a4
the 'unlock' command can be used to remove stale locks
`

func TestRemoteLockFailure(t *testing.T) {
	analysis := NewOutputAnalyser().AnalyseStringLines(ResticLockFailureOutput)

	assert.Equal(t, true, analysis.ContainsRemoteLockFailure())

	{
		since, ok := analysis.GetRemoteLockedSince()
		assert.Equal(t, true, ok)
		assert.Equal(t, time.Second+(727*time.Millisecond), since)
	}

	{
		name, ok := analysis.GetRemoteLockedBy()
		assert.Equal(t, true, ok)
		assert.Equal(t, "PID 27153 on app-server-01 by root (UID 0, GID 0)", name)
	}
}
