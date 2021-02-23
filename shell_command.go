package main

import (
	"io"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/shell"
)

type shellCommandDefinition struct {
	command  string
	args     []string
	env      []string
	useStdin bool
	stdout   io.Writer
	stderr   io.Writer
	dryRun   bool
	sigChan  chan os.Signal
	setPID   func(pid int)
}

// newShellCommand creates a new shell command definition
func newShellCommand(command string, args, env []string, dryRun bool, sigChan chan os.Signal, setPID func(pid int)) shellCommandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return shellCommandDefinition{
		command:  command,
		args:     args,
		env:      env,
		useStdin: false,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
		dryRun:   dryRun,
		sigChan:  sigChan,
		setPID:   setPID,
	}
}

// runShellCommand instantiates a shell.Command and sends the information to run the shell command
func runShellCommand(command shellCommandDefinition) (shell.Summary, error) {
	var err error

	if command.dryRun {
		clog.Infof("dry-run: %s %s", command.command, strings.Join(command.args, " "))
		return shell.Summary{}, nil
	}

	shellCmd := shell.NewSignalledCommand(command.command, command.args, command.sigChan)

	shellCmd.Stdout = command.stdout
	shellCmd.Stderr = command.stderr

	if command.useStdin {
		shellCmd.Stdin = os.Stdin
	}

	// set PID callback
	if command.setPID != nil {
		shellCmd.SetPID = command.setPID
	}

	shellCmd.Environ = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		shellCmd.Environ = append(shellCmd.Environ, command.env...)
	}

	summary, err := shellCmd.Run()
	if err != nil {
		return summary, err
	}
	return summary, nil
}
