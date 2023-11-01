package main

// ownCommandError is an error that can be returned by an own command.
// It contains an exit code that should be used to exit the program.
// There's no need to return this error if the exit code is 1.
type ownCommandError struct {
	err      error
	exitCode int
}

func (e *ownCommandError) Error() string {
	return e.err.Error()
}

func (e *ownCommandError) Unwrap() error {
	return e.err
}

func (e *ownCommandError) ExitCode() int {
	return e.exitCode
}

func newOwnCommandError(err error, exitCode int) *ownCommandError {
	return &ownCommandError{
		err:      err,
		exitCode: exitCode,
	}
}
