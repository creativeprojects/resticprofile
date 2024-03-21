//go:build !windows && !darwin

package platform

const LineSeparator = "\n"

func IsDarwin() bool {
	return false
}

func IsWindows() bool {
	return false
}

func SupportsSyslog() bool {
	return true
}

func Executable(name string) string { return name }
