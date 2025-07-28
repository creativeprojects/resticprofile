//go:build linux

package util

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// Executable returns the path name for the executable that started the current process.
// On non-Linux systems, it behaves like os.Executable.
// On Linux, it returns the path to the executable as specified in the command line arguments.
func Executable() (string, error) {
	return resolveExecutable(os.Args[0])
}

func resolveExecutable(executable string) (string, error) {
	if len(executable) == 0 {
		return "", errors.New("executable path is empty")
	}
	if executable[0] != '/' {
		// If the path is relative, we need to resolve it to an absolute path
		if executable[0] == '.' {
			wd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			// If the executable path is relative, prepend the current working directory to form an absolute path.
			executable = filepath.Join(wd, executable)
		} else {
			// If the path is not absolute, we assume it is in the PATH and resolve it.
			found, err := exec.LookPath(executable)
			if err != nil {
				return "", err
			}
			executable = filepath.Clean(found)
		}
	}
	return executable, nil
}
