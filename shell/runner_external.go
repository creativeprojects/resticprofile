package shell

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/platform"
)

type ExternalRunner struct {
	config   RunnerConfig
	shell    string
	path     string
	composer externalShellArgumentsComposer
}

func NewExternalRunner(config RunnerConfig) (*ExternalRunner, error) {
	shell := string(config.Shell)
	path, err := exec.LookPath(shell)
	if err != nil {
		return nil, err
	}
	shell = shellName(shell)
	clog.Debugf("running commands using %s at %q", shell, path)
	return &ExternalRunner{
		config:   config,
		shell:    shell,
		path:     path,
		composer: getExternalShellComposer(shell),
	}, nil
}

func (r *ExternalRunner) Run(ctx context.Context, cmdConfig CommandConfig) error {
	arguments := r.composer(r.config, cmdConfig)
	cmd := exec.CommandContext(ctx, r.path, arguments...)
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

func shellName(shell string) string {
	shell = strings.ToLower(filepath.Base(shell))

	if ext := filepath.Ext(shell); len(ext) > 0 {
		shell = shell[:len(shell)-len(ext)]
	}
	return shell
}
