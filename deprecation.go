package main

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
)

// displayDeprecationNotices displays deprecation notices for the given profile or group.
func displayDeprecationNotices(profileOrGroup any) {
	if profile, ok := profileOrGroup.(*config.Profile); ok {
		if profile.HasDeprecatedRetentionSchedule() {
			clog.Warning(`Using a schedule on a "retention" section is deprecated. Please move the schedule parameters to a "forget" section instead.`)
		}
	}
}
