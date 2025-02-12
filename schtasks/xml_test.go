//go:build !taskmaster

package schtasks

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskSchedulerIntegration(t *testing.T) {
	// some tests are using the 1st day of the month as a reference,
	// but this cause issues when we're running the tests on the first day of the month.
	// typically the test will only generate entries at a time after the time we run the test
	// for that matter let's generate a day that is not today
	dayOfTheMonth := "1"
	if time.Now().Day() == 1 {
		dayOfTheMonth = "2"
	}
	// same issue with tests on mondays
	fixedDay := "Monday"
	if time.Now().Weekday() == time.Monday {
		fixedDay = "Tuesday"
	}

	fixtures := []struct {
		description string
		schedules   []string
	}{
		{
			"only once",
			[]string{"2020-01-02 03:04"},
		},
		// daily
		{
			"once every day",
			[]string{"*-*-* 03:04"},
		},
		{
			"every hour",
			[]string{"*-*-* *:04"},
		},
		{
			"every minute",
			[]string{"*-*-* *:*"},
		},
		{
			"every minute at 12",
			[]string{"*-*-* 12:*"},
		},
		// daily - more than one
		{
			"three times a day",
			[]string{"*-*-* 03..05:04"},
		},
		{
			"twice every hour",
			[]string{"*-*-* *:04..05"},
		},
		// weekly
		{
			"once weekly",
			[]string{"mon *-*-* 03:04"},
		},
		{
			"every hour on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:04"},
		},
		{
			"every minute on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:*"},
		},
		{
			"every minute at 12 on mondays",
			[]string{"mon *-*-* 12:*"},
		},
		// more than once weekly
		{
			"twice weekly",
			[]string{"mon *-*-* 03..04:04"},
		},
		{
			"twice mondays and tuesdays",
			[]string{"mon,tue *-*-* 03:04..06"},
		},
		{
			"twice on fridays",
			[]string{"fri *-*-* *:04..05"},
		},
		// monthly
		{
			"once monthly",
			[]string{"*-01-* 03:04"},
		},
		{
			"every hour in january",
			[]string{"*-01-* *:04"},
		},
		// monthly with weekdays
		{
			"mondays in January",
			[]string{"mon *-01-* 03:04"},
		},
		{
			"every hour on Mondays in january",
			[]string{"mon *-01-* *:04"},
		},
		// some days every month
		{
			"one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03:04"},
		},
		{
			"every hour on the 1st of each month",
			[]string{"*-*-0" + dayOfTheMonth + " *:04"},
		},
		// more than once per month
		{
			"twice in one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03..04:04"},
		},
	}

	count := 0
	for _, fixture := range fixtures {
		t.Run(fixture.description, func(t *testing.T) {
			var err error
			count++
			scheduleConfig := &Config{
				ProfileName:      fmt.Sprintf("test-profile-%d", count),
				CommandName:      "test-command",
				Command:          "echo",
				Arguments:        "hello there",
				WorkingDirectory: "C:\\",
				JobDescription:   fixture.description,
			}

			schedules := make([]*calendar.Event, len(fixture.schedules))
			for index, schedule := range fixture.schedules {
				event := calendar.NewEvent()
				err := event.Parse(schedule)
				require.NoError(t, err)
				schedules[index] = event
			}
			result, sourceTask, err := createTask(scheduleConfig, schedules)
			t.Logf("result: %q\n", result)
			require.NoError(t, err)

			taskPath := getTaskPath(scheduleConfig.ProfileName, scheduleConfig.CommandName)
			taskXML, err := readTaskDefinition(taskPath)
			require.NoError(t, err)

			buffer := bytes.NewBuffer(taskXML)
			decoder := xml.NewDecoder(buffer)
			decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
				// no need for character conversion
				return input, nil
			}
			readTask := &Task{}
			err = decoder.Decode(&readTask)
			require.NoError(t, err)

			assert.Equal(t, sourceTask, readTask)

			result, err = deleteTask(taskPath)
			t.Logf("result: %q\n", result)
			require.NoError(t, err)
		})
	}
}
