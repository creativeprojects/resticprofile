package platform

import "runtime"

func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

func IsWindows() bool {
	return runtime.GOOS == "windows"
}
