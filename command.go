package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/creativeprojects/resticprofile/clog"
)

type commandDefinition struct {
	command       string
	args          []string
	env           []string
	displayStderr bool
	useStdin      bool
	shell         bool
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
		shell:         true,
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
	var cmd *exec.Cmd

	if !command.shell {
		cmd = exec.Command(command.command, command.args...)
	} else {
		flatCommand := append([]string{command.command}, command.args...)
		if runtime.GOOS == "windows" {
			shell, err := exec.LookPath("cmd.exe")
			if err != nil {
				return err
			}
			clog.Debugf("Using shell %s", shell)
			cmd = exec.Command(shell, "/C", strings.Join(flatCommand, " "))
		} else {
			shell, err := exec.LookPath("sh")
			if err != nil {
				return err
			}
			clog.Debugf("Using shell %s", shell)
			cmd = exec.Command(shell, "-c", strings.Join(flatCommand, " "))
		}
	}

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
