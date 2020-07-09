//+build darwin

package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
