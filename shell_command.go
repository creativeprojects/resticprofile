package main

import (
	"os"

	"github.com/creativeprojects/resticprofile/shell"
)

type shellCommandDefinition struct {
	command       string
	args          []string
	env           []string
	displayStderr bool
	useStdin      bool
	sigChan       chan os.Signal
}

func newShellCommand(command string, args, env []string) shellCommandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return shellCommandDefinition{
		command:       command,
		args:          args,
		env:           env,
		displayStderr: true,
		useStdin:      false,
	}
}

func runShellCommand(command shellCommandDefinition) error {
	var err error

	cmd := shell.NewSignalledCommand(command.command, command.args, command.sigChan)

	cmd.Stdout = os.Stdout
	if command.displayStderr {
		cmd.Stderr = os.Stderr
	}

	if command.useStdin {
		cmd.Stdin = os.Stdin
	}

	cmd.Environ = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		cmd.Environ = append(cmd.Environ, command.env...)
	}

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
