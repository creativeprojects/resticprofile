//go:build !windows

package shell

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterruptUnixShellCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("don't run this test in short mode")
	}
	runnerConfigs := []RunnerConfig{
		{
			DryRun: false,
			Shell:  TypeNoShell,
			Env:    []string{},
			Dir:    "",
		},
		{
			DryRun: false,
			Shell:  TypeInternalPOSIX,
			Env:    []string{},
			Dir:    "",
		},
		{
			DryRun: false,
			Shell:  TypeInternalBash,
			Env:    []string{},
			Dir:    "",
		},
		{
			DryRun: false,
			Shell:  TypeExternalPOSIX,
			Env:    []string{},
			Dir:    "",
		},
		{
			DryRun: false,
			Shell:  TypeExternalBash,
			Env:    []string{},
			Dir:    "",
		},
	}

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			stdout := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: mockBinary,
				Args:    []string{"test", "--sleep", "5000"},
				Stdout:  stdout,
			}

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(time.Second)
				cancel()
			}()
			start := time.Now()
			err = runner.Run(ctx, cmdConfig)
			require.ErrorContains(t, err, "exit status 128")

			// check it ran for more than 1s (but well under the 5s sleep - the build agent can be slow)
			duration := time.Since(start)
			assert.GreaterOrEqual(t, duration.Milliseconds(), int64(1000))
			assert.Less(t, duration.Milliseconds(), int64(4000))
		})
	}
}
