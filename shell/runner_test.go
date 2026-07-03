package shell

import (
	"bytes"
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunnerEchoCommand(t *testing.T) {
	runnerConfigs := getRunnerConfigs([]string{}, withPowershell)

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
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
			assert.Equal(t, 2, strings.Count(stdout.String(), t.Name()))
			assert.Empty(t, stderr.String())
		})
	}
}

func TestRunnerEchoEnvCommand(t *testing.T) {
	runnerConfigs := getRunnerConfigs([]string{"VAR1=value1"}, withPowershell)

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: "echo",
				Args:    []string{"$VAR1"},
				Stdin:   os.Stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			for range 2 {
				err = runner.Run(context.Background(), cmdConfig)
				require.NoError(t, err)
			}
			assert.Equal(t, 2, strings.Count(stdout.String(), "value1"))
			assert.Empty(t, stderr.String())
		})
	}
}

func TestRunnerWhoAmICommand(t *testing.T) {
	// whoami is one of the rare executable both available on unix & windows
	runnerConfigs := getRunnerConfigs(os.Environ(), withDirectRunner, withPowershell)

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: platform.Executable("whoami"),
				Args:    []string{},
				Stdin:   os.Stdin,
				Stdout:  stdout,
				Stderr:  stderr,
			}

			for range 2 {
				err = runner.Run(context.Background(), cmdConfig)
				require.NoError(t, err)
			}
			assert.Empty(t, stderr.String())
			assert.NotEmpty(t, strings.TrimSpace(stdout.String()))
		})
	}
}

func TestRunnerSetPIDCallback(t *testing.T) {
	runnerConfigs := getRunnerConfigs(os.Environ(), withDirectRunner, withPowershell)

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			called := 0
			stdout := new(bytes.Buffer)
			cmdConfig := CommandConfig{
				Command: platform.Executable("whoami"),
				Args:    []string{},
				Stdin:   os.Stdin,
				Stdout:  stdout,
				SetPID: func(pid int) {
					called++
				},
			}

			for range 2 {
				err = runner.Run(context.Background(), cmdConfig)
				require.NoError(t, err)
			}
			assert.Equal(t, 2, called)
		})
	}
}

func TestRunnerRedirectStderr(t *testing.T) {
	// powershell no longer allows >&2
	runnerConfigs := getRunnerConfigs(os.Environ())

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			bufferStdout, bufferStderr := &bytes.Buffer{}, &bytes.Buffer{}
			cmdConfig := CommandConfig{
				Command: "echo",
				Args:    []string{"error message", ">&2"},
				Stdin:   os.Stdin,
				Stdout:  bufferStdout,
				Stderr:  bufferStderr,
			}

			err = runner.Run(context.Background(), cmdConfig)
			require.NoError(t, err)

			assert.Empty(t, bufferStdout.String())
			assert.Contains(t, bufferStderr.String(), "error message")
		})
	}
}

func TestRunnerShellWorkingDir(t *testing.T) {
	runnerConfigs := getRunnerConfigs(os.Environ(), withPowershell)

	command := func(shellType Type) string {
		if platform.IsWindows() && shellType == TypeWindowsCmd {
			return "@echo %CD%"
		}
		// this is also an alias in powershell
		// https://learn.microsoft.com/en-us/powershell/scripting/learn/shell/using-aliases?view=powershell-7.6#compatibility-aliases-in-windows
		return "pwd"
	}

	for _, runnerConfig := range runnerConfigs {
		t.Run(string(runnerConfig.Shell), func(t *testing.T) {
			temp := t.TempDir()
			runnerConfig.Dir = temp

			runner, err := getRunner(runnerConfig)
			require.NoError(t, err)

			bufferStdout := &bytes.Buffer{}
			cmdConfig := CommandConfig{
				Command: command(runnerConfig.Shell),
				Args:    nil,
				Stdout:  bufferStdout,
			}

			err = runner.Run(context.Background(), cmdConfig)
			require.NoError(t, err)

			assert.Contains(t, strings.TrimSpace(bufferStdout.String()), temp)
		})
	}
}

type getRunnerConfigsOption string

const (
	withDirectRunner getRunnerConfigsOption = "with-direct-runner"
	withPowershell   getRunnerConfigsOption = "with-powershell"
)

func getRunnerConfigs(env []string, options ...getRunnerConfigsOption) []RunnerConfig {
	runnerConfigs := []RunnerConfig{
		{
			DryRun: false,
			Shell:  TypeInternalPOSIX,
			Env:    env,
			Dir:    "",
		},
		{
			DryRun: false,
			Shell:  TypeInternalBash,
			Env:    env,
			Dir:    "",
		},
	}
	if !platform.IsWindows() {
		runnerConfigs = append(runnerConfigs,
			RunnerConfig{
				DryRun: false,
				Shell:  TypeExternalPOSIX,
				Env:    env,
				Dir:    "",
			},
			RunnerConfig{
				DryRun: false,
				Shell:  TypeExternalBash,
				Env:    env,
				Dir:    "",
			},
		)
	} else {
		runnerConfigs = append(runnerConfigs,
			RunnerConfig{
				DryRun: false,
				Shell:  TypeWindowsCmd,
				Env:    env,
				Dir:    "",
			},
		)
		if slices.Contains(options, withPowershell) {
			runnerConfigs = append(runnerConfigs,
				RunnerConfig{
					DryRun: false,
					Shell:  TypeWindowsPowershell,
					Env:    env,
					Dir:    "",
				},
			)
		}
	}

	if slices.Contains(options, withDirectRunner) {
		runnerConfigs = append(runnerConfigs, RunnerConfig{
			DryRun: false,
			Shell:  TypeNoShell,
			Env:    env,
			Dir:    "",
		})
	}
	return runnerConfigs
}
