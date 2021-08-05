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
	return false
}

func IsError(err error) bool {
	return err != nil && !IsWarning(err)
}
