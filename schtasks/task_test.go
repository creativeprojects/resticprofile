//go:build windows

package schtasks

import (
	"bytes"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddLogonTrigger(t *testing.T) {
	t.Parallel()

	t.Run("for a specific user", func(t *testing.T) {
		task := NewTask()
		task.addLogonTrigger("S-1-5-21-1234")
		require.Len(t, task.Triggers.LogonTrigger, 1)
		assert.Equal(t, "S-1-5-21-1234", task.Triggers.LogonTrigger[0].UserId)

		buffer := &bytes.Buffer{}
		require.NoError(t, createTaskFile(task, buffer))
		assert.Contains(t, buffer.String(), "<LogonTrigger>")
		assert.Contains(t, buffer.String(), "<UserId>S-1-5-21-1234</UserId>")
	})

	t.Run("for any user", func(t *testing.T) {
		task := NewTask()
		task.addLogonTrigger("")
		require.Len(t, task.Triggers.LogonTrigger, 1)

		buffer := &bytes.Buffer{}
		require.NoError(t, createTaskFile(task, buffer))
		assert.Contains(t, buffer.String(), "<LogonTrigger>")
		// an empty UserId must be omitted so the trigger fires for any user logon
		assert.NotContains(t, buffer.String(), "<UserId>")
	})

	t.Run("coexists with time triggers", func(t *testing.T) {
		task := NewTask()
		task.addTimeTrigger(time.Date(2020, 1, 2, 3, 4, 0, 0, time.UTC))
		task.addLogonTrigger("user")
		assert.Len(t, task.Triggers.TimeTrigger, 1)
		assert.Len(t, task.Triggers.LogonTrigger, 1)
	})
}

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
	for i := range 31 {
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
