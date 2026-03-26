//go:build windows

package schtasks

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompileDifferences(t *testing.T) {
	testData := []struct {
		input       string
		differences []time.Duration
		unique      []time.Duration
	}{
		{
			"1..4,6,8:00",
			[]time.Duration{1 * time.Hour, 1 * time.Hour, 1 * time.Hour, 2 * time.Hour, 2 * time.Hour},
			[]time.Duration{1 * time.Hour, 2 * time.Hour},
		},
		{
			"Sat,Sun 0,12:00",
			[]time.Duration{12 * time.Hour},
			[]time.Duration{12 * time.Hour},
		},
		{
			"mon *-11..12-* 1,13:00",
			[]time.Duration{12 * time.Hour},
			[]time.Duration{12 * time.Hour},
		},
	}

	for _, testItem := range testData {
		event := calendar.NewEvent()
		err := event.Parse(testItem.input)
		require.NoError(t, err)
		ref, err := time.Parse(time.ANSIC, "Mon Jan 2 12:00:00 2006")
		require.NoError(t, err)
		start := event.Next(ref)
		diff, uniques := compileDifferences(event.GetAllInBetween(start, start.Add(24*time.Hour)))
		assert.ElementsMatch(t, testItem.differences, diff, "duration between triggers")
		assert.ElementsMatch(t, testItem.unique, uniques, "unique set of durations between triggers")
	}
}

func TestConvertDaysOfMonth(t *testing.T) {
	allDays := make([]int, 31)
	for i := 0; i <= 30; i++ {
		allDays[i] = i + 1
	}

	testData := []struct {
		description string
		input       []int
		expected    DaysOfMonth
	}{
		{
			description: "empty input returns all 31 days",
			input:       []int{},
			expected:    DaysOfMonth{Day: allDays},
		},
		{
			description: "nil input returns all 31 days",
			input:       nil,
			expected:    DaysOfMonth{Day: allDays},
		},
		{
			description: "single day",
			input:       []int{15},
			expected:    DaysOfMonth{Day: []int{15}},
		},
		{
			description: "first day of month",
			input:       []int{1},
			expected:    DaysOfMonth{Day: []int{1}},
		},
		{
			description: "last day of month",
			input:       []int{31},
			expected:    DaysOfMonth{Day: []int{31}},
		},
		{
			description: "multiple specific days",
			input:       []int{1, 15, 28},
			expected:    DaysOfMonth{Day: []int{1, 15, 28}},
		},
		{
			description: "consecutive days",
			input:       []int{10, 11, 12, 13},
			expected:    DaysOfMonth{Day: []int{10, 11, 12, 13}},
		},
	}

	for _, testItem := range testData {
		t.Run(testItem.description, func(t *testing.T) {
			result := convertDaysOfMonth(testItem.input)
			assert.Equal(t, testItem.expected, result)
		})
	}
}
