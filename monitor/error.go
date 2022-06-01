package monitor

import (
	"errors"
	"os/exec"

	"github.com/creativeprojects/resticprofile/constants"
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
		return exitErr.ExitCode() == constants.ExitCodeWarning
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
