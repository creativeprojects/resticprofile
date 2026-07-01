package shell

import (
	"context"
	"os/exec"
)

type Internal struct {
	command string
	args    *Args
}

func NewInternal(command string, args *Args) *Internal {
	return &Internal{
		command: command,
		args:    args,
	}
}

func (i *Internal) Start(ctx context.Context) (*exec.Cmd, error) {
	args := i.args.GetAll()
	cmd := exec.CommandContext(ctx, i.command, args...)
	err := cmd.Start()
	return cmd, err
}
