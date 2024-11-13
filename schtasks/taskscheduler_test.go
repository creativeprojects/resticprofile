//go:build windows

package schtasks

import (
	"bytes"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/rickb777/date/period"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionWeekdaysToBitmap(t *testing.T) {
	testData := []struct {
		weekdays []int
		bitmap   uint16
	}{
		{nil, 0},
		{[]int{}, 0},
		{[]int{0}, 1},
		{[]int{1}, 2},
		{[]int{2}, 4},
		{[]int{7}, 1},
		{[]int{1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{0, 1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{0, 1, 2, 3, 4, 5, 6}, 127},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.bitmap, convertWeekdaysToBitmap(testItem.weekdays))
	}
}

func TestConversionMonthsToBitmap(t *testing.T) {
	testData := []struct {
		months []int
		bitmap uint16
	}{
		{nil, 0},
		{[]int{}, 4095}, // every month
		{[]int{0}, 0},
		{[]int{1}, 1},
		{[]int{2}, 2},
		{[]int{7}, 64},
		{[]int{1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{0, 1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{1, 2, 3, 4, 5, 6}, 63},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.bitmap, convertMonthsToBitmap(testItem.months))
	}
}

func TestConversionDaysToBitmap(t *testing.T) {
	testData := []struct {
		days   []int
		bitmap uint32
	}{
		{nil, 0},
		{[]int{}, math.MaxInt32}, // every day
		{[]int{0}, 0},
		{[]int{1}, 1},
		{[]int{2}, 2},
		{[]int{7}, 64},
		{[]int{1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{0, 1, 2, 3, 4, 5, 6, 7}, 127},
		{[]int{1, 2, 3, 4, 5, 6}, 63},
	}

	for _, testItem := range testData {
		assert.Equal(t, testItem.bitmap, convertDaysToBitmap(testItem.days))
	}
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

func TestTaskSchedulerConversion(t *testing.T) {
	testData := []string{
		"2020-01-01",
		"*:0,15,30,45",
		"sat,sun 3:30",
		"*-*-1",
		"mon *-1..10-*",
	}
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	schedules := make([]*calendar.Event, len(testData))
	for index, testEvent := range testData {
		event := calendar.NewEvent()
		err := event.Parse(testEvent)
		require.NoError(t, err)
		schedules[index] = event
	}
	task := taskmaster.Definition{}
	createSchedules(&task, schedules)

	require.Len(t, task.Triggers, 5)

	// 1st task should be a single event
	singleEvent, ok := task.Triggers[0].(taskmaster.TimeTrigger)
	require.True(t, ok)
	assert.Equal(t, "2020-01-01 00:00:00", singleEvent.StartBoundary.Format("2006-01-02 15:04:05"))

	// 2nd task will be a daily recurring
	dailyEvent, ok := task.Triggers[1].(taskmaster.DailyTrigger)
	require.True(t, ok)
	assert.Equal(t, period.NewHMS(0, 15, 0), dailyEvent.RepetitionInterval)  // 15 minutes
	assert.Equal(t, period.NewHMS(23, 45, 0), dailyEvent.RepetitionDuration) // 23h 45 minutes

	// 3rd task will be a weekly recurring
	weeklyEvent, ok := task.Triggers[2].(taskmaster.WeeklyTrigger)
	require.True(t, ok)
	assert.Equal(t, getWeekdayBit(int(time.Saturday))+getWeekdayBit(int(time.Sunday)), uint16(weeklyEvent.DaysOfWeek))

	// 4th task will be a monthly recurring
	monthlyEvent, ok := task.Triggers[3].(taskmaster.MonthlyTrigger)
	require.True(t, ok)
	t.Logf("%+v", monthlyEvent)

	// 5th task will be a monthly with day of week recurring
	monthlyDOWEvent, ok := task.Triggers[4].(taskmaster.MonthlyDOWTrigger)
	require.True(t, ok)
	t.Logf("%+v", monthlyDOWEvent)
}

func TestStatusUnknownTask(t *testing.T) {
	err := Connect()
	defer Close()
	assert.NoError(t, err)

	err = Status("test", "test")
	assert.Error(t, err)
	t.Log(err)
}

// this creates tasks through the API and compares it with the XML export
func TestCreationOfTasks(t *testing.T) {
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
		everyMonth += `<` + time.Month(month).String() + ` />\s*`
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
			`<TimeTrigger>\s*<StartBoundary>2020-01-02T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*</TimeTrigger>`,
			1,
		},
		// daily
		{
			"once every day",
			[]string{"*-*-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour",
			[]string{"*-*-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT23H</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute",
			[]string{"*-*-* *:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>P1D</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute at 12",
			[]string{"*-*-* 12:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T12:\d{2}:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT59M</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		// daily - more than one
		{
			"three times a day",
			[]string{"*-*-* 03..05:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT2H</Duration>\s*</Repetition>\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			1,
		},
		{
			"twice every hour",
			[]string{"*-*-* *:04..05"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByDay>\s*<DaysInterval>1</DaysInterval>\s*</ScheduleByDay>\s*</CalendarTrigger>`,
			48,
		},
		// weekly
		{
			"once weekly",
			[]string{"mon *-*-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT23H</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<` + fixedDay + ` />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute on mondays",
			[]string{strings.ToLower(fixedDay)[:3] + " *-*-* *:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>P1D</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<` + fixedDay + ` />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every minute at 12 on mondays",
			[]string{"mon *-*-* 12:*"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T12:\d{2}:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT59M</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		// more than once weekly
		{
			"twice weekly",
			[]string{"mon *-*-* 03..04:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1H</Interval>\s*<Duration>PT1H</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"twice mondays and tuesdays",
			[]string{"mon,tue *-*-* 03:04..06"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<Repetition>\s*<Interval>PT1M</Interval>\s*<Duration>PT2M</Duration>\s*</Repetition>\s*<ScheduleByWeek>\s*<WeeksInterval>1</WeeksInterval>\s*<DaysOfWeek>\s*<Monday />\s*<Tuesday />\s*</DaysOfWeek>\s*</ScheduleByWeek>\s*</CalendarTrigger>`,
			1,
		},
		// monthly
		{
			"once monthly",
			[]string{"*-01-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonth>\s*<Months>\s*<January />\s*</Months>\s*<DaysOfMonth>\s*` + everyDay + `</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour in january",
			[]string{"*-01-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonth>\s*<Months>\s*<January />\s*</Months>\s*<DaysOfMonth>\s*` + everyDay + `</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			24,
		},
		// monthly with weekdays
		{
			"mondays in January",
			[]string{"mon *-01-* 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonthDayOfWeek>\s*<Months>\s*<January />\s*</Months>\s*<Weeks>\s*<Week>1</Week>\s*<Week>2</Week>\s*<Week>3</Week>\s*<Week>4</Week>\s*<Week>Last</Week>\s*</Weeks>\s*<DaysOfWeek>\s*<Monday />\s*</DaysOfWeek>\s*</ScheduleByMonthDayOfWeek>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on Mondays in january",
			[]string{"mon *-01-* *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-01-\d{2}T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonthDayOfWeek>\s*<Months>\s*<January />\s*</Months>\s*<Weeks>\s*<Week>1</Week>\s*<Week>2</Week>\s*<Week>3</Week>\s*<Week>4</Week>\s*<Week>Last</Week>\s*</Weeks>\s*<DaysOfWeek>\s*<Monday />\s*</DaysOfWeek>\s*</ScheduleByMonthDayOfWeek>\s*</CalendarTrigger>`,
			24,
		},
		// some days every month
		{
			"one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			1,
		},
		{
			"every hour on the 1st of each month",
			[]string{"*-*-0" + dayOfTheMonth + " *:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			24, // 1 per hour
		},
		// more than once per month
		{
			"twice in one day per month",
			[]string{"*-*-0" + dayOfTheMonth + " 03..04:04"},
			`<CalendarTrigger>\s*<StartBoundary>\d{4}-\d{2}-0` + dayOfTheMonth + `T\d{2}:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*<ScheduleByMonth>\s*<Months>\s*` + everyMonth + `</Months>\s*<DaysOfMonth>\s*<Day>` + dayOfTheMonth + `</Day>\s*</DaysOfMonth>\s*</ScheduleByMonth>\s*</CalendarTrigger>`,
			2,
		},
	}

	count := 0
	for _, fixture := range fixtures {
		count++
		t.Run(fixture.description, func(t *testing.T) {
			err := Connect()
			defer Close()
			assert.NoError(t, err)

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
			// user logged in doesn't need a password
			err = createUserLoggedOnTask(scheduleConfig, schedules)
			assert.NoError(t, err)
			defer func() {
				_ = Delete(scheduleConfig.ProfileName, scheduleConfig.CommandName)
			}()

			taskName := getTaskPath(scheduleConfig.ProfileName, scheduleConfig.CommandName)
			buffer, err := exportTask(taskName)
			assert.NoError(t, err)

			pattern := regexp.MustCompile(fixture.expected)
			match := pattern.FindAllString(buffer, -1)
			assert.Len(t, match, fixture.expectedMatchCount)

			if t.Failed() {
				t.Log(buffer)
			}
		})
	}
}

func exportTask(taskName string) (string, error) {
	buffer := &bytes.Buffer{}
	cmd := exec.Command("schtasks", "/query", "/xml", "/tn", taskName)
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.String(), err
}

func TestRegisteredTasks(t *testing.T) {
	tasks := []Config{
		{
			ProfileName:      "test1",
			CommandName:      "backup",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test1",
		},
		{
			ProfileName:      "test 2",
			CommandName:      "check",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test 2",
		},
		{
			ProfileName:      "test 3",
			CommandName:      "forget",
			Command:          "echo",
			Arguments:        "hello there",
			WorkingDirectory: "C:\\",
			JobDescription:   "test 3",
		},
	}
	err := Connect()
	defer Close()
	assert.NoError(t, err)

	event := calendar.NewEvent()
	err = event.Parse("2020-01-02 03:04") // will never get triggered
	require.NoError(t, err)

	for _, task := range tasks {
		// user logged in doesn't need a password
		err = createUserLoggedOnTask(&task, []*calendar.Event{event})
		assert.NoError(t, err)

		defer func() {
			_ = Delete(task.ProfileName, task.CommandName)
		}()
	}

	registeredTasks, err := Registered()
	assert.NoError(t, err)

	assert.ElementsMatch(t, tasks, registeredTasks)
}
