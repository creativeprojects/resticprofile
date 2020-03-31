package constants

import "github.com/creativeprojects/resticprofile/priority"

var (
	PriorityValues = map[string]int{
		"idle":       priority.Idle,
		"background": priority.Background,
		"low":        priority.Low,
		"normal":     priority.Normal,
		"high":       priority.High,
		"highest":    priority.Highest,
	}
)
