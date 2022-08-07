//go:build windows

package platform

func IsDarwin() bool {
	return false
}

func IsWindows() bool {
	return true
}

func SupportsSyslog() bool {
	return false
}
