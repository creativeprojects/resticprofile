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

func convertMonths(input []int) Months {
	if len(input) == 0 {
		return Months{
			January:   Month,
			February:  Month,
			March:     Month,
			April:     Month,
			May:       Month,
			June:      Month,
			July:      Month,
			August:    Month,
			September: Month,
			October:   Month,
			November:  Month,
			December:  Month,
		}
	}
	var months Months
	for _, month := range input {
		switch month {
		case 1:
			months.January = Month
		case 2:
			months.February = Month
		case 3:
			months.March = Month
		case 4:
			months.April = Month
		case 5:
			months.May = Month
		case 6:
			months.June = Month
		case 7:
			months.July = Month
		case 8:
			months.August = Month
		case 9:
			months.September = Month
		case 10:
			months.October = Month
		case 11:
			months.November = Month
		case 12:
			months.December = Month
		}
	}
	return months
}

func convertDaysOfMonth(input []int) DaysOfMonth {
	if len(input) == 0 {
		all := make([]int, 31)
		for i := 1; i <= 31; i++ {
			all[i-1] = i
		}
		return DaysOfMonth{all}
	}
	return DaysOfMonth{input}
}

func convertWeekdays(input []int) DaysOfWeek {
	var weekDays DaysOfWeek
	if len(input) == 0 {
		return weekDays
	}
	for _, weekday := range input {
		switch weekday {
		case 0, 7:
			weekDays.Sunday = WeekDay
		case 1:
			weekDays.Monday = WeekDay
		case 2:
			weekDays.Tuesday = WeekDay
		case 3:
			weekDays.Wednesday = WeekDay
		case 4:
			weekDays.Thursday = WeekDay
		case 5:
			weekDays.Friday = WeekDay
		case 6:
			weekDays.Saturday = WeekDay
		}
	}
	return weekDays
}
