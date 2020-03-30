package main

import (
	"os"
	"os/exec"
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

func runCommands(commands []commandDefinition) error {
	for _, command := range commands {
		err := runCommand(command)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCommand(command commandDefinition) error {
	cmd := exec.Command(command.command, command.args...)

	cmd.Stdout = os.Stdout
	if command.displayStderr {
		cmd.Stderr = os.Stderr
	}

	if command.useStdin {
		cmd.Stdin = os.Stdin
	}

	cmd.Env = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		cmd.Env = append(cmd.Env, command.env...)
	}

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
