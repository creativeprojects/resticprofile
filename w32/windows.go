// +build windows

package w32

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/creativeprojects/resticprofile/constants"
	"golang.org/x/sys/windows"
)

const (
	ATTACH_PARENT_PROCESS windows.Handle = 0x0ffffffff
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

// AttachParentConsole detach from the current console and attach to the parent process console
func AttachParentConsole() error {
	err := FreeConsole()
	if err != nil {
		return err
	}
	err = AttachConsole(ATTACH_PARENT_PROCESS)
	if err != nil {
		return err
	}
	return nil
}

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

	err := windows.ShellExecute(GetConsoleWindow(), verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}
	return nil
}

// GetConsoleWindow from windows kernel32 API
func GetConsoleWindow() windows.Handle {
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	ret, _, _ := getConsoleWindow.Call()
	return windows.Handle(ret)
}

// FreeConsole from windows kernel32 API
func FreeConsole() error {
	freeConsole := kernel32.NewProc("FreeConsole")
	ret, _, _ := freeConsole.Call()
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

// AttachConsole from windows kernel32 API
func AttachConsole(consoleOwner windows.Handle) error {
	attachConsole := kernel32.NewProc("AttachConsole")
	ret, _, _ := attachConsole.Call(uintptr(consoleOwner))
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}
