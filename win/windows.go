//go:build windows

package win

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"golang.org/x/sys/windows"
)

// RunElevated restart resticprofile in elevated mode.
// the parameter is the port where the parent is listening
func RunElevated(port int) error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := fmt.Sprintf("--%s --%s %d %s",
		constants.FlagAsChild,
		constants.FlagPort,
		port,
		parseArguments(os.Args[1:]),
	)

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	clog.Debugf("starting command \"%s %s %s\", current directory = %q", verb, exe, args, cwd)

	err := windows.ShellExecute(GetConsoleWindow(), verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}
	return nil
}

// GetConsoleWindow from windows kernel32 API
func GetConsoleWindow() windows.Handle {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	ret, _, _ := getConsoleWindow.Call()
	return windows.Handle(ret)
}

// parseArguments takes a slice of strings as input and returns a single string.
// It processes each argument, and if an argument contains a space, it wraps it in double quotes.
// Finally, it joins all the processed arguments into a single string separated by spaces.
func parseArguments(args []string) string {
	output := make([]string, len(args))
	for i, arg := range args {
		if strings.Contains(arg, " ") {
			output[i] = `"` + arg + `"`
			continue
		}
		output[i] = arg
	}
	return strings.Join(output, " ")
}
