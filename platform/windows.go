//go:build windows

package platform

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
