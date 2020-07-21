//+build windows

package win

import (
	"testing"

	"github.com/capnspacehook/taskmaster"
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
func TestTaskSchedulerConversion(t *testing.T) {
	testData := []string{
		"2020-01-01",
		"*:0,15,30,45",
	}
	schedules := make([]*calendar.Event, len(testData))
	for index, testEvent := range testData {
		event := calendar.NewEvent()
		err := event.Parse(testEvent)
		require.NoError(t, err)
		schedules[index] = event
	}
	task := taskmaster.Definition{}
	taskScheduler := NewTaskScheduler(&config.ScheduleConfig{})
	taskScheduler.createSchedules(&task, schedules)
	// first task should be a single event
	singleEvent, ok := task.Triggers[0].(taskmaster.TimeTrigger)
	require.True(t, ok)
	assert.Equal(t, "2020-01-01 00:00:00", singleEvent.StartBoundary.Format("2006-01-02 15:04:05"))
	t.Logf("%+v", task.Triggers[1])
	// second task will be a daily recurring
	dailyEvent, ok := task.Triggers[1].(taskmaster.DailyTrigger)
	require.True(t, ok)
	assert.Equal(t, period.NewHMS(0, 15, 0), dailyEvent.RepetitionInterval)
	assert.Equal(t, period.NewYMD(0, 0, 1), dailyEvent.RepetitionDuration)
}
