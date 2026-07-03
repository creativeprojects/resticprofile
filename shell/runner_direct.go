package shell

import (
	"context"
	"errors"
	"os/exec"
	"syscall"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/platform"
)

type DirectRunner struct {
	config RunnerConfig
}

func NewDirectRunner(config RunnerConfig) *DirectRunner {
	clog.Debug("running commands directly (not through a shell)")
	return &DirectRunner{
		config: config,
	}
}

func (r *DirectRunner) Run(ctx context.Context, cmdConfig CommandConfig) error {
	cmd := exec.CommandContext(ctx, cmdConfig.Command, cmdConfig.Args...)
	cmd.Dir = r.config.Dir
	cmd.Stdin = cmdConfig.Stdin
	cmd.Stdout = cmdConfig.Stdout
	cmd.Stderr = cmdConfig.Stderr
	cmd.Env = r.config.Env
	cmd.Cancel = nil

	if !platform.IsWindows() {
		// register a cascade cancellation from the context
		cmd.Cancel = func() error {
			if cmd.Process == nil {
				return errors.New("nil process")
			}
			return cmd.Process.Signal(syscall.SIGINT)
		}
	}

	// spawn the child process
	if err := cmd.Start(); err != nil {
		return err
	}
	if cmdConfig.SetPID != nil {
		// send the PID back (to write down in a lockfile)
		cmdConfig.SetPID(cmd.Process.Pid)
	}

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
