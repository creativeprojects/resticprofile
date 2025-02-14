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
