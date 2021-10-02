//go:build windows
// +build windows

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
		strings.Join(os.Args[1:], " "),
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
