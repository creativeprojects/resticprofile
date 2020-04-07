package constants

import "github.com/creativeprojects/resticprofile/priority"

var (
	// PriorityValues is the map between the name and the value
	PriorityValues = map[string]int{
		"idle":       priority.Idle,
		"background": priority.Background,
		"low":        priority.Low,
		"normal":     priority.Normal,
		"high":       priority.High,
		"highest":    priority.Highest,
	}
)
