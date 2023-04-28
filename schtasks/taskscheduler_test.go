//go:build windows

package schtasks

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/capnspacehook/taskmaster"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/rickb777/date/period"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversionWeekdaysToBitmap(t *testing.T) {
	testData := []struct {
		weekdays []int
		bitmap   int
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
		assert.ElementsMatch(t, testItem.differences, diff)
		assert.ElementsMatch(t, testItem.unique, uniques)
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

	// 1st task should be a single event
	singleEvent, ok := task.Triggers[0].(taskmaster.TimeTrigger)
	require.True(t, ok)
	assert.Equal(t, "2020-01-01 00:00:00", singleEvent.StartBoundary.Format("2006-01-02 15:04:05"))

	// 2nd task will be a daily recurring
	dailyEvent, ok := task.Triggers[1].(taskmaster.DailyTrigger)
	require.True(t, ok)
	assert.Equal(t, period.NewHMS(0, 15, 0), dailyEvent.RepetitionInterval)
	assert.Equal(t, period.NewYMD(0, 0, 1), dailyEvent.RepetitionDuration)

	// 3rd task will be a weekly recurring
	weeklyEvent, ok := task.Triggers[2].(taskmaster.WeeklyTrigger)
	require.True(t, ok)
	assert.Equal(t, getWeekdayBit(int(time.Saturday))+getWeekdayBit(int(time.Sunday)), int(weeklyEvent.DaysOfWeek))

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

func TestCreationOfTasks(t *testing.T) {
	fixtures := []struct {
		description string
		schedules   []string
		expected    string
	}{
		{
			"single date time",
			[]string{"2020-01-02 03:04"},
			`<TimeTrigger>\s*<StartBoundary>2020-01-02T03:04:00</StartBoundary>\s*(<ExecutionTimeLimit>PT0S</ExecutionTimeLimit>)?\s*</TimeTrigger>`,
		},
	}

	count := 0
	for _, fixture := range fixtures {
		count++
		t.Run(fixture.description, func(t *testing.T) {
			err := Connect()
			defer Close()
			assert.NoError(t, err)

			scheduleConfig := &config.ScheduleConfig{
				Title:            "test",
				SubTitle:         strconv.Itoa(count),
				Command:          "echo",
				Arguments:        []string{"hello"},
				WorkingDirectory: "C:\\",
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
			defer Delete(scheduleConfig.Title, scheduleConfig.SubTitle)

			taskName := getTaskPath(scheduleConfig.Title, scheduleConfig.SubTitle)
			buffer, err := exportTask(taskName)
			assert.NoError(t, err)

			pattern := regexp.MustCompile(fixture.expected)
			match := pattern.FindString(buffer)
			assert.NotEmptyf(t, match, "expected to find %q in %q", fixture.expected, buffer)
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
