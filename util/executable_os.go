//go:build darwin || windows

package util

import "os"

// Executable returns the path name for the executable that started the current process.
// On Darwin and Windows, it behaves like os.Executable.
// On Linux and BSDs, it returns the path to the executable as specified in the command line arguments.
func Executable() (string, error) {
	return os.Executable()
}
