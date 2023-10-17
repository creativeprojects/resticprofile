package main

import (
	"errors"
	"fmt"
	"strings"
)

type commandError struct {
	scd    shellCommandDefinition
	stderr string
	err    error
}

func newCommandError(command shellCommandDefinition, stderr string, err error) *commandError {
	return &commandError{
		scd:    command,
		stderr: stderr,
		err:    err,
	}
}

func (c *commandError) Error() string {
	return c.err.Error()
}

func (c *commandError) Unwrap() error {
	return c.err
}

func (c *commandError) Commandline() string {
	args := ""
	argsList := c.scd.publicArgs
	if len(argsList) > 0 {
		args = fmt.Sprintf(" \"%s\"", strings.Join(argsList, "\" \""))
	}
	return fmt.Sprintf("\"%s\"%s", c.scd.command, args)
}

func (c *commandError) Stderr() string {
	return c.stderr
}

func (c *commandError) ExitCode() (int, error) {
	if exitError, ok := asExitError(c.err); ok {
		return exitError.ExitCode(), nil
	}
	return 0, errors.New("exit code not available")
}
