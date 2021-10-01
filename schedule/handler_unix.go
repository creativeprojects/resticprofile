//go:build !darwin && !windows

package schedule

import "github.com/creativeprojects/resticprofile/constants"

func NewHandler(config SchedulerConfig) Handler {
	if config.Type() == constants.SchedulerCrond {
		return NewHandlerCrond(config)
	}
	return NewHandlerSystemd(config)
}
