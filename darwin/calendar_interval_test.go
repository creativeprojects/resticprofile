//go:build darwin

package darwin

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
)

func TestParseCalendarIntervals(t *testing.T) {
	tests := []struct {
		name      string
		intervals []CalendarInterval
		expected  []string
	}{
		{
			name: "Single interval",
			intervals: []CalendarInterval{
				{
					intervalMinute:  30,
					intervalHour:    14,
					intervalWeekday: 3,
					intervalDay:     15,
					intervalMonth:   6,
				},
			},
			expected: []string{"Wed *-06-15 14:30:00"},
		},
		{
			name: "Multiple intervals",
			intervals: []CalendarInterval{
				{
					intervalMinute: 0,
				},
				{
					intervalMinute: 30,
				},
			},
			expected: []string{"*-*-* *:00,30:00"},
		},
		{
			name:      "Empty intervals",
			intervals: []CalendarInterval{},
			expected:  []string{"*-*-* *:*:00"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCalendarIntervals(tt.intervals)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCalendarIntervalsFromScheduleTree(t *testing.T) {
	testData := []struct {
		input    string
		expected []CalendarInterval
	}{
		{"*-*-*", []CalendarInterval{
			{"Hour": 0, "Minute": 0},
		}},
		{"*:0,30", []CalendarInterval{
			{"Minute": 0},
			{"Minute": 30},
		}},
		{"0,12:20", []CalendarInterval{
			{"Hour": 0, "Minute": 20},
			{"Hour": 12, "Minute": 20},
		}},
		{"0,12:20,40", []CalendarInterval{
			{"Hour": 0, "Minute": 20},
			{"Hour": 0, "Minute": 40},
			{"Hour": 12, "Minute": 20},
			{"Hour": 12, "Minute": 40},
		}},
		{"Mon..Fri *-*-* *:0,30:00", []CalendarInterval{
			{"Weekday": 1, "Minute": 0},
			{"Weekday": 1, "Minute": 30},
			{"Weekday": 2, "Minute": 0},
			{"Weekday": 2, "Minute": 30},
			{"Weekday": 3, "Minute": 0},
			{"Weekday": 3, "Minute": 30},
			{"Weekday": 4, "Minute": 0},
			{"Weekday": 4, "Minute": 30},
			{"Weekday": 5, "Minute": 0},
			{"Weekday": 5, "Minute": 30},
		}},
		// First sunday of the month at 3:30am
		{"Sun *-*-01..06 03:30:00", []CalendarInterval{
			{"Day": 1, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 2, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 3, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 4, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 5, "Weekday": 0, "Hour": 3, "Minute": 30},
			{"Day": 6, "Weekday": 0, "Hour": 3, "Minute": 30},
		}},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			event := calendar.NewEvent()
			err := event.Parse(testItem.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, testItem.expected, getCalendarIntervalsFromScheduleTree(generateTreeOfSchedules(event)))
		})
	}
}
