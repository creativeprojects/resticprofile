package schtasks

import (
	"bytes"
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerCreationFromXML(t *testing.T) {
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
	everyDay := ""
	for day := 1; day <= 31; day++ {
		everyDay += `<Day>` + strconv.Itoa(day) + `</Day>\s*`
	}
	everyMonth := ""
	for month := 1; month <= 12; month++ {
		everyMonth += `<` + time.Month(month).String() + `></` + time.Month(month).String() + `>\s*`
	}

	fixtures := []struct {
		description        string
		schedules          []string
		expected           string
		expectedMatchCount int
	}{
		{
			"only once",
			[]string{"2020-01-02 03:04"},
			`<TimeTrigger>\s*<StartBoundary>2020-01-02T03:04:00Z</StartBoundary>\s*</TimeTrigger>`,
			1,
		},
		// daily
		{
			"once every day",
			[]string{"*-*-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00Z</StartBoundary>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour",
			[]string{"*-*-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:04:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT23H</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute",
			[]string{"*-*-* *:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>P1D</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute at 12",
			[]string{"*-*-* 12:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T12:\d{2}:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT59M</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		// daily - more than one
		{
			"three times a day",
			[]string{"*-*-* 03..05:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT2H</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"twice every hour",
			[]string{"*-*-* *:04..05"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z</StartBoundary>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			48,
		},
		// weekly
		{
			"once weekly",
			[]string{"mon *-*-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00Z</StartBoundary>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday></Monday>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:04:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT23H</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<` + fixedDay + `></` + fixedDay + `>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>P1D</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<` + fixedDay + `></` + fixedDay + `>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute at 12 on mondays",
			[]string{"mon *-*-* 12:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T12:\d{2}:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT59M</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday></Monday>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		// more than once weekly
		{
			"twice weekly",
			[]string{"mon *-*-* 03..04:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT1H</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday></Monday>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"twice mondays and tuesdays",
			[]string{"mon,tue *-*-* 03:04..06"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00Z</StartBoundary>\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT2M</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday></Monday>\s*<Tuesday></Tuesday>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"twice on fridays",
			[]string{"fri *-*-* *:04..05"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00Z</StartBoundary>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Friday></Friday>\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			48,
		},
		// monthly
		{
			"once monthly",
			[]string{"*-01-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T03:04:00Z</StartBoundary>\s*<ScheduleByMonth>\s*<Months>\s*<January></January>\s*</Months>\s*<DaysOfMonth>\s*` + everyDay + `</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour in january",
			[]string{"*-01-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T\d{2}:04:00Z</StartBoundary>\s*<ScheduleByMonth>\s*<Months>\s*<January></January>\s*</Months>\s*<DaysOfMonth>\s*` + everyDay + `</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			24,
		},
		// monthly with weekdays
		{
			"mondays in January",
			[]string{"mon *-01-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T03:04:00Z</StartBoundary>\s*<ScheduleByMonthDayOfWeek>\s*<Months>\s*<January></January>\s*</Months>\s*<Weeks>\s*<Week>1</Week>\s*<Week>2</Week>\s*<Week>3</Week>\s*<Week>4</Week>\s*<Week>Last</Week>\s*</Weeks>\s*<DaysOfWeek>\s*<Monday></Monday>\s*</DaysOfWeek>\s*</ScheduleByMonthDayOfWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on Mondays in january",
			[]string{"mon *-01-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T\d{2}:04:00Z</StartBoundary>\s*<ScheduleByMonthDayOfWeek>\s*<Months>\s*<January></January>\s*</Months>\s*<Weeks>\s*<Week>1</Week>\s*<Week>2</Week>\s*<Week>3</Week>\s*<Week>4</Week>\s*<Week>Last</Week>\s*</Weeks>\s*<DaysOfWeek>\s*<Monday></Monday>\s*</DaysOfWeek>\s*</ScheduleByMonthDayOfWeek>\s*</CalendarTrigger>`,
			24,
		},
		// // some days every month
		{
			"one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T03:04:00Z</StartBoundary>\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on the 1st of each month",
			[]string{"*-*-0" + dayOfTheMonth + " *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T\d{2}:04:00Z</StartBoundary>\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			24, // 1 per hour
		},
		// // more than once per month
		{
			"twice in one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03..04:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T\d{2}:04:00Z</StartBoundary>\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			2,
		},
	}

	count := 0
	for _, fixture := range fixtures {
		count++
		t.Run(fixture.description, func(t *testing.T) {
			var err error
			scheduleConfig := &Config{
				ProfileName:      "test",
				CommandName:      strconv.Itoa(count),
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
			buffer := &bytes.Buffer{}
			err = createTaskFile(scheduleConfig, schedules, buffer)
			require.NoError(t, err)

			pattern := regexp.MustCompile(fixture.expected)
			match := pattern.FindAll(buffer.Bytes(), -1)
			assert.Len(t, match, fixture.expectedMatchCount, fixture.expected)

			if t.Failed() {
				t.Log(buffer.String())
			}
		})
	}
}

func createTaskFile(config *Config, schedules []*calendar.Event, w io.Writer) error {
	var err error
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	task := NewTask()
	task.RegistrationInfo.Description = config.JobDescription
	task.AddExecAction(ExecAction{
		Command:          config.Command,
		Arguments:        config.Arguments,
		WorkingDirectory: config.WorkingDirectory,
	})
	task.AddSchedules(schedules)
	err = encoder.Encode(&task)
	if err != nil {
		return err
	}
	return err
}
