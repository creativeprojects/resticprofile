package progress

import (
	"errors"
	"os/exec"
)

func IsSuccess(err error) bool {
	return err == nil
}

func IsWarning(err error) bool {
	if err == nil {
		return false
	}
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == 3
	}
	// so far, internal warning is only used in unit tests
	warn := &InternalWarning{}
	if errors.As(err, &warn) {
		return true
	}
	return false
}

func IsError(err error) bool {
	return err != nil && !IsWarning(err)
}

type InternalWarning struct {
}

func (w InternalWarning) Error() string {
	return "internal warning"
}
