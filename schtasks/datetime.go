package schtasks

import (
	"time"

	"github.com/rickb777/date/period"
)

// compileDifferences is creating two slices: the first one is the duration between each trigger,
// the second one is a list of all the differences in between
//
// Example:
//
//	input = 01:00, 02:00, 03:00, 04:00, 06:00, 08:00
//	first list = 1H, 1H, 1H, 2H, 2H
//	second list = 1H, 2H
func compileDifferences(recurrences []time.Time) ([]time.Duration, []time.Duration) {
	// now calculate the difference in between each
	differences := make([]time.Duration, len(recurrences)-1)
	for i := 0; i < len(recurrences)-1; i++ {
		differences[i] = recurrences[i+1].Sub(recurrences[i])
	}
	// check if they're all the same
	compactDifferences := make([]time.Duration, 0, len(differences))
	var previous time.Duration = 0
	for _, difference := range differences {
		if difference.Seconds() != previous.Seconds() {
			compactDifferences = append(compactDifferences, difference)
			previous = difference
		}
	}
	return differences, compactDifferences
}

func getRepetionDuration(start time.Time, recurrences []time.Time) period.Period {
	last := recurrences[len(recurrences)-1]
	duration := period.Between(start, last)
	// convert 1439 minutes to 23 hours
	if duration.DurationApprox() == 1439*time.Minute {
		duration = period.NewHMS(0, 1440, 0)
	}
	return duration
}
