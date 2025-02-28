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
		return exitErr.ExitCode() == constants.ResticExitCodeWarning
	}
	// so far, internal warning is only used in unit tests
	warn := &InternalWarningError{}
	return errors.As(err, &warn)
}

func IsError(err error) bool {
	return err != nil && !IsWarning(err)
}

type InternalWarningError struct {
}

func (w InternalWarningError) Error() string {
	return "internal warning"
}
