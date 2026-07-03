//go:build windows

package shell

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterruptWindowsShellCommand(t *testing.T) {
	clog.SetTestLog(t)

	runnerConfigs := []RunnerConfig{
		// {
		// 	DryRun: false,
		// 	Shell:  TypeNoShell,
		// 	Env:    []string{},
		// 	Dir:    "",
		// },
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
		// {
		// 	DryRun: false,
		// 	Shell:  TypeWindowsCmd,
		// 	Env:    []string{},
		// 	Dir:    "",
		// },
		// {
		// 	DryRun: false,
		// 	Shell:  TypeWindowsPowershell,
		// 	Env:    []string{},
		// 	Dir:    "",
		// },
	}

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			// will need to be moved into the internal runner code
			binary := strings.ReplaceAll(mockBinary, `\`, "/")

			output := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: binary,
				Args:    []string{"test", "--sleep 1000"},
				Stdout:  output,
				Stderr:  output,
			}

			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				time.Sleep(500 * time.Millisecond)
				cancel()
			}()
			start := time.Now()
			err = runner.Run(ctx, cmdConfig)

			assert.Equal(t, "", output.String())
			require.NoError(t, err)

			// check it ran for more than a second, meaning it wasn't cancelled
			duration := time.Since(start)
			assert.GreaterOrEqual(t, duration.Milliseconds(), int64(1000))
		})
	}
}
