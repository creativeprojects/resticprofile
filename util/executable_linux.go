//go:build linux

package util

import (
	"errors"
	"os"
	"path/filepath"
)

// Executable returns the path name for the executable that started the current process.
// On non-Linux systems, it behaves like os.Executable.
// On Linux, it returns the path to the executable as specified in the command line arguments.
func Executable() (string, error) {
	executable := os.Args[0]
	if len(executable) == 0 {
		return "", errors.New("executable path is empty")
	}
	if executable[0] != '/' {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		// If the executable path is relative, prepend the current working directory to form an absolute path.
		executable = filepath.Join(wd, executable)
	}
	return executable, nil
}
