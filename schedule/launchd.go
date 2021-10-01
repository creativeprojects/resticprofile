//go:build darwin

package schedule

import (
	"github.com/creativeprojects/resticprofile/constants"
)

// LaunchJob is an agent definition for launchd
type LaunchdJob struct {
	Label                 string             `plist:"Label"`
	Program               string             `plist:"Program"`
	ProgramArguments      []string           `plist:"ProgramArguments"`
	EnvironmentVariables  map[string]string  `plist:"EnvironmentVariables,omitempty"`
	StandardInPath        string             `plist:"StandardInPath,omitempty"`
	StandardOutPath       string             `plist:"StandardOutPath,omitempty"`
	StandardErrorPath     string             `plist:"StandardErrorPath,omitempty"`
	WorkingDirectory      string             `plist:"WorkingDirectory"`
	StartInterval         int                `plist:"StartInterval,omitempty"`
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval,omitempty"`
	ProcessType           string             `plist:"ProcessType"`
	LowPriorityIO         bool               `plist:"LowPriorityIO"`
	Nice                  int                `plist:"Nice"`
}

var priorityValues = map[string]string{
	constants.SchedulePriorityBackground: "Background",
	constants.SchedulePriorityStandard:   "Standard",
}
