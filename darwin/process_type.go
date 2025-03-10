//go:build darwin

package darwin

import "github.com/creativeprojects/resticprofile/constants"

type ProcessType string

const (
	ProcessTypeBackground ProcessType = "Background"
	ProcessTypeStandard   ProcessType = "Standard"
)

func NewProcessType(schedulePriority string) ProcessType {
	switch schedulePriority {
	case constants.SchedulePriorityBackground:
		return ProcessTypeBackground

	case constants.SchedulePriorityStandard:
		return ProcessTypeStandard

	default:
		return ""
	}
}
