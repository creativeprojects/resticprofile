package main

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
)

func displayProfileDeprecationNotices(profile *config.Profile) {
	if profile.HasDeprecatedRetentionSchedule() {
		clog.Warning(`Using a schedule on a "retention" section is deprecated. Please move the schedule parameters to a "forget" section instead.`)
	}
}
