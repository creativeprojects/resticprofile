package shell

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ResticLockFailureOutput = `
repo already locked, waiting up to 2m1s115.3ms for the lock
unable to create lock in backend: repository is already locked by PID 27153 on app-server-01 by root (UID 0, GID 0)
lock was created at 2021-04-17 19:00:16 (1s727.5ms ago)
storage ID 870530a4
the 'unlock' command can be used to remove stale locks
`

func TestRemoteLockFailure(t *testing.T) {
	t.Parallel()

	analysis := NewOutputAnalyser()
	require.NoError(t, analysis.AnalyseStringLines(ResticLockFailureOutput))

	assert.Equal(t, true, analysis.ContainsRemoteLockFailure())

	t.Run("GetRemoteLockedSince", func(t *testing.T) {
		since, ok := analysis.GetRemoteLockedSince()
		assert.Equal(t, true, ok)
		assert.Equal(t, time.Second+(727*time.Millisecond), since)
	})

	t.Run("GetRemoteLockedBy", func(t *testing.T) {
		name, ok := analysis.GetRemoteLockedBy()
		assert.Equal(t, true, ok)
		assert.Equal(t, "PID 27153 on app-server-01 by root (UID 0, GID 0)", name)
	})

	t.Run("GetRemoteLockedMaxWait", func(t *testing.T) {
		wait, ok := analysis.GetRemoteLockedMaxWait()
		assert.Equal(t, true, ok)
		assert.Equal(t, 2*time.Minute+time.Second+(115*time.Millisecond), wait)
	})
}

func TestCustomErrorCallback(t *testing.T) {
	t.Parallel()

	var analyser *OutputAnalyser
	invoked := 0
	var cbError error
	triggerLine := "--TRIGGER_CALLBACK--"

	init := func(t *testing.T, minCount, maxCalls int, stopOnError bool) error {
		analyser = NewOutputAnalyser()
		invoked = 0
		cbError = nil
		return analyser.SetCallback("cb-test", ".+TRIGGER_CALLBACK.+", minCount, maxCalls, stopOnError, func(line string) error {
			require.Equal(t, line, triggerLine)
			invoked++
			return cbError
		})
	}

	writeTrigger := func(t *testing.T) {
		require.NoError(t, analyser.AnalyseStringLines(triggerLine))
	}

	t.Run("RequiresMinCountForEachCall", func(t *testing.T) {
		require.NoError(t, init(t, 3, 4, false))

		for c := 0; c < 2; c++ {
			for i := 0; i < 3; i++ {
				assert.Equal(t, c, invoked)
				writeTrigger(t)
			}
			assert.Equal(t, c+1, invoked)
		}
	})

	t.Run("CanLimitMaxCalls", func(t *testing.T) {
		require.NoError(t, init(t, 0, 4, false))

		for i, c := 0, 0; c < 8; c++ {
			assert.Equal(t, i, invoked)
			writeTrigger(t)
			if i < 4 {
				i++
			}
		}
	})

	t.Run("CallbackErrorIsReturned", func(t *testing.T) {
		require.NoError(t, init(t, 0, 0, true))
		cbError = fmt.Errorf("cb-error")
		assert.ErrorIs(t, analyser.AnalyseStringLines(triggerLine), cbError)
	})

	t.Run("CallbackErrorCanBeSkipped", func(t *testing.T) {
		require.NoError(t, init(t, 0, 0, false))
		cbError = fmt.Errorf("cb-error")
		assert.ErrorIs(t, analyser.AnalyseStringLines(triggerLine), nil)
	})

	t.Run("RegexErrorIsReturned", func(t *testing.T) {
		require.NoError(t, init(t, 0, 0, false))
		err := analyser.SetCallback("fail", "[", 0, 0, false, nil)
		assert.EqualError(t, err, "error parsing regexp: missing closing ]: `[`")
	})

	t.Run("CanUnregisterCallback", func(t *testing.T) {
		require.NoError(t, init(t, 0, 0, false))
		writeTrigger(t)
		assert.Equal(t, 1, invoked)

		require.NoError(t, analyser.SetCallback("cb-test", ".", 0, 0, false, nil))
		writeTrigger(t)
		assert.Equal(t, 1, invoked)
	})
}
