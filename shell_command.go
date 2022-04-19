package main

import (
	"io"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/progress"
	"github.com/creativeprojects/resticprofile/shell"
)

type shellCommandDefinition struct {
	command    string
	args       []string
	publicArgs []string
	env        []string
	stdin      io.ReadCloser
	stdout     io.Writer
	stderr     io.Writer
	dryRun     bool
	sigChan    chan os.Signal
	setPID     shell.SetPID
	scanOutput shell.ScanOutput
}

// newShellCommand creates a new shell command definition
func newShellCommand(command string, args, env []string, dryRun bool, sigChan chan os.Signal, setPID func(pid int)) shellCommandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return shellCommandDefinition{
		command:    command,
		args:       args,
		publicArgs: args,
		env:        env,
		stdin:      nil,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		dryRun:     dryRun,
		sigChan:    sigChan,
		setPID:     setPID,
	}
}

// runShellCommand instantiates a shell.Command and sends the information to run the shell command
func runShellCommand(command shellCommandDefinition) (progress.Summary, string, error) {
	var err error

	if command.dryRun {
		clog.Infof("dry-run: %s %s", command.command, strings.Join(command.publicArgs, " "))
		return progress.Summary{}, "", nil
	}

	shellCmd := shell.NewSignalledCommand(command.command, command.args, command.sigChan)

	shellCmd.Stdout = command.stdout
	shellCmd.Stderr = command.stderr

	if command.stdin != nil {
		shellCmd.Stdin = command.stdin
	}

	// set PID callback
	if command.setPID != nil {
		shellCmd.SetPID = command.setPID
	}

	shellCmd.Environ = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		shellCmd.Environ = append(shellCmd.Environ, command.env...)
	}

	// scan output
	if command.scanOutput != nil {
		shellCmd.ScanStdout = command.scanOutput
	}

	summary, stderr, err := shellCmd.Run()
	if err != nil {
		return summary, stderr, err
	}
	return summary, stderr, nil
}
