package crond

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEvents(t *testing.T) {
	testData := []struct {
		cron     string
		expected string
	}{
		{"00 00 * * 1,2,3,4,6,0", "Sun..Thu,Sat *-*-* 00:00:00"},
		{"23 01,02 * * 1,0", "Sun,Mon *-*-* 01,02:23:00"},
		{"00 00 01 * 3", "Wed *-*-01 00:00:00"},
		{"48 17 * * 3", "Wed *-*-* 17:48:00"},
		{"02 01 15 10 2,3,4,5,6", "Tue..Sat *-10-15 01:02:00"},
		{"00 00 07 * *", "*-*-07 00:00:00"},
		{"00 00 15 10 *", "*-10-15 00:00:00"},
		{"00 17 * 12 1", "Mon *-12-* 17:00:00"},
		{"00 17 * 12 0", "Sun *-12-* 17:00:00"},
		{"30 * 01-03 * 1,5", "Mon,Fri *-*-01..03 *:30:00"},
		{"10,20,30 12-14 * * *", "*-*-* 12..14:10,20,30:00"},
		{"05 08 05 03 *", "*-03-05 08:05:00"},
		{"05 08 * * *", "*-*-* 08:05:00"},
		{"40 05 * * *", "*-*-* 05:40:00"},
		{"05 08 05 12 6,0", "Sun,Sat *-12-05 08:05:00"},
		{"05 08 * * 6,0", "Sun,Sat *-*-* 08:05:00"},
		{"40 05 05 03 *", "*-03-05 05:40:00"},
		{"00 00 05 02-04 *", "*-02..04-05 00:00:00"},
		{"00 00 05 03 *", "*-03-05 00:00:00"},
		{"00 00 * * 1,2,3,4,5,6,0", "Sun..Sat *-*-* 00:00:00"},
		{"00 00 * * 0,1", "Sun,Mon *-*-* 00:00:00"},
		{"00\t00 * * 0,1", "Sun,Mon *-*-* 00:00:00"},   // should replace tab by space
		{"00 00    * * 0,1", "Sun,Mon *-*-* 00:00:00"}, // should compact all spaces into one
	}

	for _, testRun := range testData {
		t.Run(testRun.cron, func(t *testing.T) {
			event, err := parseEvent(testRun.cron)
			require.NoError(t, err)
			assert.Equal(t, testRun.expected, event.String())
		})
	}
}

func TestFailingParseEvents(t *testing.T) {
	testData := []struct {
		cron string
	}{
		{""},
		{" "},
		{"  "},
		{"   "},
		{"    "},
		{"     "},
		{"      "},
		{"       "},
		{"        "},
		{"         "},
		{"99 00 * * 0,1"},
		{"0- 00 * * 0,1"},
		{"-0 00 * * 0,1"},
		{"0, 00 * * 0,1"},
		{",0 00 * * 0,1"},
		{"invalid"},
	}

	for _, testRun := range testData {
		t.Run(testRun.cron, func(t *testing.T) {
			_, err := parseEvent(testRun.cron)
			require.Error(t, err)
		})
	}
}
