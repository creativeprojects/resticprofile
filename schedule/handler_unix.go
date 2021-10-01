//go:build !darwin && !windows

package schedule

import "github.com/creativeprojects/resticprofile/constants"

// NewHandler creates a crond or systemd handler depending on the configuration
func NewHandler(config SchedulerConfig) Handler {
	if config.Type() == constants.SchedulerCrond {
		return NewHandlerCrond(config)
	}
	return NewHandlerSystemd(config)
}
