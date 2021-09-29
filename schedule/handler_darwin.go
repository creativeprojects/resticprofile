//go:build darwin
// +build darwin

package schedule

type HandlerLaunchd struct {
	//
}

// Available verifies launchd is available on this system
func (h *HandlerLaunchd) Available() error {
	return lookupBinary("launchd", launchdBin)
}

var (
	_ Handler = &HandlerLaunchd{}
)
