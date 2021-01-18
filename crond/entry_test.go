//+build !darwin,!windows

package crond

import (
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyUserEvent(t *testing.T) {
	entry := NewEntry(calendar.NewEvent(), "", "", "", "command line")
	buffer := &strings.Builder{}
	err := entry.Generate(buffer)
	require.NoError(t, err)
	assert.Equal(t, "* * * * *\tcommand line\n", buffer.String())
}

func TestEvents(t *testing.T) {
	testData := []struct {
		event    string
		expected string
	}{
		{"Sat,Thu,Mon..Wed,Sat..Sun", "00 00 * * 1,2,3,4,6,0"},
		{"Mon,Sun 12-*-* 2,1:23", "23 01,02 * * 1,0"},
		{"Wed *-1", "00 00 01 * 3"},
		{"Wed..Wed,Wed *-1", "00 00 01 * 3"},
		{"Wed, 17:48", "48 17 * * 3"},
		{"Wed..Sat,Tue 12-10-15 1:2:3", "02 01 15 10 2,3,4,5,6"},
		{"*-*-7 0:0:0", "00 00 07 * *"},
		{"10-15", "00 00 15 10 *"},
		{"monday *-12-* 17:00", "00 17 * 12 1"},
		{"sunday *-12-* 17:00", "00 17 * 12 0"},
		{"Mon,Fri *-*-3,1,2 *:30:45", "30 * 01-03 * 1,5"},
		{"12,14,13,12:20,10,30", "10,20,30 12-14 * * *"},
		{"12..14:10,20,30", "10,20,30 12-14 * * *"},
		{"03-05 08:05:40", "05 08 05 03 *"},
		{"08:05:40", "05 08 * * *"},
		{"05:40", "40 05 * * *"},
		{"Sat,Sun 12-05 08:05:40", "05 08 05 12 6,0"},
		{"Sat,Sun 08:05:40", "05 08 * * 6,0"},
		{"2003-03-05 05:40", "40 05 05 03 *"},
		{"2003-02..04-05", "00 00 05 02-04 *"},
		{"2003-03-05", "00 00 05 03 *"},
		{"03-05", "00 00 05 03 *"},
		{"hourly", "00 * * * *"},
		{"daily", "00 00 * * *"},
		{"monthly", "00 00 01 * *"},
		{"weekly", "00 00 * * 1"},
		{"yearly", "00 00 01 01 *"},
		{"annually", "00 00 01 01 *"},
		{"mon..sun", "00 00 * * 1,2,3,4,5,6,0"},
		{"sun..mon", "00 00 * * 0,1"},
	}

	for _, testRun := range testData {
		t.Run(testRun.event, func(t *testing.T) {
			event := calendar.NewEvent()
			err := event.Parse(testRun.event)
			require.NoError(t, err)

			entry := NewEntry(event, "", "", "", "command line")
			buffer := &strings.Builder{}
			err = entry.Generate(buffer)
			require.NoError(t, err)
			assert.Equal(t, testRun.expected+"\tcommand line\n", buffer.String())
		})
	}
}
