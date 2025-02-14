//go:build darwin

package schedule

import (
	"testing"

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
			result := parseCalendarIntervals(tt.intervals)
			assert.Equal(t, tt.expected, result)
		})
	}
}
