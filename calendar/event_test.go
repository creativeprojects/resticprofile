package calendar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventKeywords(t *testing.T) {
	testData := []struct{ keyword, expected string }{
		{"minutely", "*-*-* *:*:00"},
		{"hourly", "*-*-* *:00:00"},
		{"daily", "*-*-* 00:00:00"},
		{"monthly", "*-*-01 00:00:00"},
		{"weekly", "Mon *-*-* 00:00:00"},
		{"yearly", "*-01-01 00:00:00"},
		{"quarterly", "*-01,04,07,10-01 00:00:00"},
		{"semiannually", "*-01,07-01 00:00:00"},
	}

	for _, testItem := range testData {
		t.Run(testItem.keyword, func(t *testing.T) {
			event := NewEvent(specialKeywords[testItem.keyword])
			assert.Equal(t, testItem.expected, event.String())
		})
	}
}

func TestEventParse(t *testing.T) {
	// Not all forms of systemd.time are allowed (yet?)
	// Commented out are the examples that are not valid in our implementation
	testData := []struct{ input, expected string }{
		{"Sat,Thu,Mon..Wed,Sat..Sun", "Mon..Thu,Sat..Sun *-*-* 00:00:00"},
		{"Mon,Sun 12-*-* 2,1:23", "Mon,Sun 2012-*-* 01,02:23:00"},
		{"Wed *-1", "Wed *-*-01 00:00:00"},
		{"Wed..Wed,Wed *-1", "Wed *-*-01 00:00:00"},
		{"Wed, 17:48", "Wed *-*-* 17:48:00"},
		{"Wed..Sat,Tue 12-10-15 1:2:3", "Tue..Sat 2012-10-15 01:02:03"},
		{"*-*-7 0:0:0", "*-*-07 00:00:00"},
		{"10-15", "*-10-15 00:00:00"},
		{"monday *-12-* 17:00", "Mon *-12-* 17:00:00"},
		{"Mon,Fri *-*-3,1,2 *:30:45", "Mon,Fri *-*-01..03 *:30:45"},
		{"12,14,13,12:20,10,30", "*-*-* 12..14:10,20,30:00"},
		{"12..14:10,20,30", "*-*-* 12..14:10,20,30:00"},
		// {"mon,fri *-1/2-1,3 *:30:45", "Mon,Fri *-01/2-01,03 *:30:45"},
		{"03-05 08:05:40", "*-03-05 08:05:40"},
		{"08:05:40", "*-*-* 08:05:40"},
		{"05:40", "*-*-* 05:40:00"},
		{"Sat,Sun 12-05 08:05:40", "Sat,Sun *-12-05 08:05:40"},
		{"Sat,Sun 08:05:40", "Sat,Sun *-*-* 08:05:40"},
		{"2003-03-05 05:40", "2003-03-05 05:40:00"},
		// {"05:40:23.4200004/3.1700005", "*-*-* 05:40:23.420000/3.170001"},
		{"2003-02..04-05", "2003-02..04-05 00:00:00"},
		// {"2003-03-05 05:40 UTC", "2003-03-05 05:40:00 UTC"},
		{"2003-03-05", "2003-03-05 00:00:00"},
		{"03-05", "*-03-05 00:00:00"},
		{"hourly", "*-*-* *:00:00"},
		{"daily", "*-*-* 00:00:00"},
		// {"daily UTC", "*-*-* 00:00:00 UTC"},
		{"monthly", "*-*-01 00:00:00"},
		{"weekly", "Mon *-*-* 00:00:00"},
		// {"weekly Pacific/Auckland", "Mon *-*-* 00:00:00 Pacific/Auckland"},
		{"yearly", "*-01-01 00:00:00"},
		{"annually", "*-01-01 00:00:00"},
		// {"*:2/3", "*-*-* *:02/3:00"},
	}

	for _, testItem := range testData {
		t.Run(testItem.input, func(t *testing.T) {
			event := NewEvent()
			err := event.Parse(testItem.input)
			assert.NoError(t, err)
			assert.Equal(t, testItem.expected, event.String())
		})
	}
}

func TestParseInvalidEvents(t *testing.T) {
	testData := []string{
		"",
		"u",
		"u..mon",
		"mon..u",
		"u-u",
		"13-01",
		"1-32",
		"1-",
		"-1",
		"1:",
		":1",
		"1:99",
		"24:2",
		"1:2:60",
	}

	for _, testItem := range testData {
		t.Run(testItem, func(t *testing.T) {
			event := NewEvent()
			err := event.Parse(testItem)
			assert.Error(t, err)
			t.Log(err)
		})
	}
}
