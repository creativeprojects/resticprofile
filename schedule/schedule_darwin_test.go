//+build darwin

package schedule

import (
	"bytes"
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

func TestGetCombinationItemsFromCalendarValues(t *testing.T) {
	// this is a bit much to run, just to create 1 weekday and 1 minute values
	// we might need to refactor if the setup is using too much code
	schedule := calendar.NewEvent()
	err := schedule.Parse("Mon..Fri *-*-* *:0,30:00")
	require.NoError(t, err)
	//

	fields := []*calendar.Value{
		schedule.WeekDay,
		schedule.Month,
		schedule.Day,
		schedule.Hour,
		schedule.Minute,
	}

	// create list of permutable items
	total, items := getCombinationItemsFromCalendarValues(fields)
	assert.Equal(t, 10, total)
	assert.Len(t, items, 7)
}

func TestConvertCombinationToCalendarInterval(t *testing.T) {
	testData := [][]combinationItem{
		{
			{calendar.TypeMinute, 15},
			{calendar.TypeMinute, 45},
		},
		{
			{calendar.TypeWeekDay, 1},
			{calendar.TypeWeekDay, 7},
		},
	}
	output := convertCombinationToCalendarInterval(testData)
	t.Logf("%+v", output)
}

func TestPListEncoderWithCalendarInterval(t *testing.T) {
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>Day</key><integer>1</integer><key>Hour</key><integer>0</integer></dict></plist>`
	entry := newCalendarInterval()
	setCalendarIntervalValueFromType(entry, 1, calendar.TypeDay)
	setCalendarIntervalValueFromType(entry, 0, calendar.TypeHour)
	buffer := &bytes.Buffer{}
	encoder := plist.NewEncoder(buffer)
	err := encoder.Encode(entry)
	require.NoError(t, err)
	assert.Equal(t, expected, buffer.String())
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
