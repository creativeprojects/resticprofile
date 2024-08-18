package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/creativeprojects/resticprofile/shell"
)

type shellCommandDefinition struct {
	command     string
	args        []string
	publicArgs  []string
	env         []string
	shell       []string
	dir         string
	stdin       io.ReadCloser
	stdout      io.Writer
	stderr      io.Writer
	dryRun      bool
	sigChan     chan os.Signal
	setPID      shell.SetPID
	scanOutput  shell.ScanOutput
	streamError []config.StreamErrorSection
}

// newShellCommand creates a new shell command definition
func newShellCommand(command string, args, env, shell []string, dryRun bool, sigChan chan os.Signal, setPID func(pid int)) shellCommandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return shellCommandDefinition{
		command:    command,
		args:       args,
		publicArgs: args,
		env:        env,
		shell:      shell,
		stdin:      nil,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		dryRun:     dryRun,
		sigChan:    sigChan,
		setPID:     setPID,
	}
}

// runShellCommand instantiates a shell.Command and sends the information to run the shell command
func runShellCommand(command shellCommandDefinition) (summary monitor.Summary, stderr string, err error) {
	if command.dryRun {
		clog.Infof("dry-run: %s %s", command.command, strings.Join(command.publicArgs, " "))
	}

	shellCmd := shell.NewSignalledCommand(command.command, command.args, command.sigChan)

	shellCmd.Shell = command.shell
	shellCmd.Stdout = command.stdout
	shellCmd.Stderr = command.stderr

	if command.dryRun {
		shellBinary, args, commandErr := shellCmd.GetShellCommand()
		if commandErr != nil {
			clog.Warningf("command error: %s", commandErr.Error())
		} else {
			// The following line may send confidential values to log (combination of --trace --dry-run).
			clog.Tracef("dry-run shell: %s %s", shellBinary, strings.Join(args, " "))
		}
		return
	}

	if command.stdin != nil {
		shellCmd.Stdin = command.stdin
	}

	// set PID callback
	if command.setPID != nil {
		shellCmd.SetPID = command.setPID
	}

	shellCmd.Environ = os.Environ()
	if len(command.env) > 0 {
		shellCmd.Environ = append(shellCmd.Environ, command.env...)
	}

	// If Dir is the empty string, Run runs the command in the
	// calling process's current directory.
	shellCmd.Dir = command.dir

	// scan output
	if command.scanOutput != nil {
		shellCmd.ScanStdout = command.scanOutput
	}

	// stderr callbacks
	if err = setupStreamErrorHandlers(&command, shellCmd); err != nil {
		return
	}

	summary, stderr, err = shellCmd.Run()
	return
}

func setupStreamErrorHandlers(command *shellCommandDefinition, shellCmd *shell.Command) error {
	for i, e := range command.streamError {
		commandLine := e.Run

		callback := func(line string) error {
			errorCmd := shell.NewSignalledCommand(commandLine, nil, command.sigChan)
			errorCmd.Environ = shellCmd.Environ
			errorCmd.Shell = command.shell
			errorCmd.Stdout = command.stdout
			errorCmd.Stderr = command.stderr

			clog.Debugf(`starting stream error callback "%s" for line "%s"`, errorCmd.Command, line)
			errorCmd.Stdin = strings.NewReader(line)

			_, stderr, err := errorCmd.Run()
			if err != nil {
				err = newCommandError(shellCommandDefinition{publicArgs: []string{commandLine}}, stderr, err)
			}
			return err
		}

		err := shellCmd.OnErrorCallback(fmt.Sprintf("%d", i+1), e.Pattern, e.MinMatches, e.MaxRuns, callback)
		if err != nil {
			err = fmt.Errorf("stream error callback: %s failed to register %s: %w", e.Run, e.Pattern, err)
			return err
		}
	}
	return nil
}
