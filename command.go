package main

import (
	"os"

	"github.com/creativeprojects/resticprofile/shell"
)

type commandDefinition struct {
	command       string
	args          []string
	env           []string
	displayStderr bool
	useStdin      bool
}

func newCommand(command string, args, env []string) commandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return commandDefinition{
		command:       command,
		args:          args,
		env:           env,
		displayStderr: true,
		useStdin:      false,
	}
}

func runCommand(command commandDefinition) error {
	var err error

	cmd := shell.NewCommand(command.command, command.args)

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

func runShellCommand(command string) error {
	var err error

	cmd := shell.NewCommand(command, nil)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
