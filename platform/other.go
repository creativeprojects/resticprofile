//go:build !windows && !darwin

package platform

func IsDarwin() bool {
	return false
}

func IsWindows() bool {
	return false
}

func SupportsSyslog() bool {
	return true
}
