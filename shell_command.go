package main

import (
	"io"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/shell"
)

type shellCommandDefinition struct {
	command       string
	args          []string
	env           []string
	displayStderr bool
	useStdin      bool
	stdout        io.Writer
	dryRun        bool
	sigChan       chan os.Signal
}

// newShellCommand creates a new shell command definition
func newShellCommand(command string, args, env []string, dryRun bool, sigChan chan os.Signal) shellCommandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return shellCommandDefinition{
		command:       command,
		args:          args,
		env:           env,
		displayStderr: true,
		useStdin:      false,
		stdout:        os.Stdout,
		dryRun:        dryRun,
		sigChan:       sigChan,
	}
}

// runShellCommand instantiates a shell.Command and sends the information to run the shell command
func runShellCommand(command shellCommandDefinition) error {
	var err error

	if command.dryRun {
		clog.Infof("dry-run: %s %s", command.command, strings.Join(command.args, " "))
		return nil
	}

	shellCmd := shell.NewSignalledCommand(command.command, command.args, command.sigChan)

	shellCmd.Stdout = command.stdout
	if command.displayStderr {
		shellCmd.Stderr = os.Stderr
	}

	if command.useStdin {
		shellCmd.Stdin = os.Stdin
	}

	shellCmd.Environ = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		shellCmd.Environ = append(shellCmd.Environ, command.env...)
	}

	err = shellCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
