//go:build windows

package platform

import "fmt"

const LineSeparator = "\r\n"

func IsDarwin() bool {
	return false
}

func IsWindows() bool {
	return true
}

func SupportsSyslog() bool {
	return false
}

func Executable(name string) string { return fmt.Sprintf("%s.exe", name) }
