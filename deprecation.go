package main

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
)

// displayDeprecationNotices displays deprecation notices for the given profile or group.
func displayDeprecationNotices(profileOrGroup config.Schedulable) {
	if profile, ok := profileOrGroup.(interface{ HasDeprecatedRetentionSchedule() bool }); ok {
		if profile.HasDeprecatedRetentionSchedule() {
			clog.Warning(`Using a schedule on a "retention" section is deprecated. Please move the schedule parameters to a "forget" section instead.`)
		}
	}
}
