package shell

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerEchoCommand(t *testing.T) {
	runnerConfigs := []RunnerConfig{
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
	}
	if platform.IsWindows() {
		runnerConfigs = append(runnerConfigs,
			RunnerConfig{
				DryRun: false,
				Shell:  TypeWindowsCmd,
				Env:    []string{},
				Dir:    "",
			},
			RunnerConfig{
				DryRun: false,
				Shell:  TypeWindowsPowershell,
				Env:    []string{},
				Dir:    "",
			},
		)
	} else {
		runnerConfigs = append(runnerConfigs,
			RunnerConfig{
				DryRun: false,
				Shell:  TypeExternalPOSIX,
				Env:    []string{},
				Dir:    "",
			},
			RunnerConfig{
				DryRun: false,
				Shell:  TypeExternalBash,
				Env:    []string{},
				Dir:    "",
			},
		)
	}

	for _, runnerConfig := range runnerConfigs {
		name := string(runnerConfig.Shell)
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: "echo",
				Args:    []string{t.Name()},
				Stdin:   os.Stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			for range 2 {
				err = runner.Run(context.Background(), cmdConfig)
				require.NoError(t, err)
			}
			expected := t.Name() + "\n" + t.Name() + "\n"
			assert.Equal(t, expected, strings.ReplaceAll(stdout.String(), "\r\n", "\n"))
			assert.Empty(t, stderr.String())
		})
	}
}
