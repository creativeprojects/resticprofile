package shell

import (
	"context"
	"strings"

	"github.com/creativeprojects/clog"
)

type DryRunner struct {
}

func NewDryRunner() *DryRunner {
	clog.Debug("using dry-run shell")
	return &DryRunner{}
}

func (r *DryRunner) Run(ctx context.Context, cmd CommandConfig) error {
	clog.Infof("dry-run: %s %s", cmd.Command, strings.Join(cmd.PublicArgs, " "))
	return nil
}
