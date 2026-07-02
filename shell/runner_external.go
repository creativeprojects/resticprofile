package shell

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/creativeprojects/clog"
)

type ExternalRunner struct {
	shell string
	path  string
}

func NewExternalRunner(config RunnerConfig) (*ExternalRunner, error) {
	shell := string(config.Shell)
	path, err := exec.LookPath(shell)
	if err != nil {
		return nil, err
	}
	clog.Debugf("running commands using %s at %q", filepath.Base(shell), path)
	return &ExternalRunner{
		shell: shell,
		path:  path,
	}, nil
}

func (r *ExternalRunner) Run(ctx context.Context, cmd CommandConfig) error {
	return nil
}
