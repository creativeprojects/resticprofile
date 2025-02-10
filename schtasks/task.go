package schtasks

import (
	"encoding/xml"
	"os/user"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/rickb777/date/period"
)

const (
	dateFormat  = "2006-01-02T15:04:05-07:00"
	maxTriggers = 60
	author      = "Author"
	tasksPath   = `\resticprofile backup\`
	taskSchema  = "http://schemas.microsoft.com/windows/2004/02/mit/task"
)

type RegistrationInfo struct {
	Date        string `xml:"Date"`
	Author      string `xml:"Author"`
	Description string `xml:"Description"`
	URI         string `xml:"URI"`
}

type Task struct {
	XMLName          xml.Name         `xml:"Task"`
	Version          string           `xml:"version,attr"`
	Xmlns            string           `xml:"xmlns,attr"`
	RegistrationInfo RegistrationInfo `xml:"RegistrationInfo"`
	Triggers         Triggers         `xml:"Triggers"`
	Principals       Principals       `xml:"Principals"`
	Settings         Settings         `xml:"Settings"`
	Actions          Actions          `xml:"Actions"`
}

func NewTask() Task {
	var userID string
	if currentUser, err := user.Current(); err == nil {
		userID = currentUser.Uid
	}
	task := Task{
		XMLName: xml.Name{Space: taskSchema, Local: "Task"},
		Version: "1.2",
		Xmlns:   taskSchema,
		RegistrationInfo: RegistrationInfo{
			Date:   time.Now().Format(dateFormat),
			Author: constants.ApplicationName,
		},
		Principals: Principals{
			Principal: Principal{
				ID:        author,
				UserId:    userID,
				LogonType: LogonTypeInteractiveToken,
				RunLevel:  RunLevelDefault,
			},
		},
		Settings: Settings{
			Compatibility:              TaskCompatibilityAT,
			DisallowStartIfOnBatteries: true,
			// Enabled:                    true,
			IdleSettings: IdleSettings{
				Duration:      period.NewHMS(0, 10, 0), // PT10M
				WaitTimeout:   period.NewHMS(1, 0, 0),  // PT1H
				StopOnIdleEnd: true,
			},
			MultipleInstancesPolicy: MultipleInstancesIgnoreNew,
			Priority:                8,
			StopIfGoingOnBatteries:  true,
			// ExecutionTimeLimit:      period.NewHMS(72, 0, 0), // PT72H
		},
		Actions: Actions{
			Context: author,
		},
	}
	return task
}

// AddExecAction returns the same instance of Task (for chaining)
func (t *Task) AddExecAction(action ExecAction) *Task {
	if t.Actions.Exec == nil {
		t.Actions.Exec = []ExecAction{action}
		return t
	}
	t.Actions.Exec = append(t.Actions.Exec, action)
	return t
}

func (t *Task) AddSchedules(schedules []*calendar.Event) {
	for _, schedule := range schedules {
		if triggerOnce, ok := schedule.AsTime(); ok {
			// one time only
			t.addTimeTrigger(triggerOnce)
			continue
		}
		if schedule.IsDaily() {
			// recurring daily
			t.addDailyTrigger(schedule)
			continue
		}
		if schedule.IsWeekly() {
			t.addWeeklyTrigger(schedule)
			continue
		}
		if schedule.IsMonthly() {
			t.addMonthlyTrigger(schedule)
			continue
		}
		clog.Warningf("cannot convert schedule '%s' into a task scheduler equivalent", schedule.String())
	}
}

func (t *Task) addTimeTrigger(triggerOnce time.Time) {
	timeTrigger := TimeTrigger{
		StartBoundary: triggerOnce.Format(dateFormat),
	}
	if t.Triggers.TimeTrigger == nil {
		t.Triggers.TimeTrigger = []TimeTrigger{timeTrigger}
		return
	}
	t.Triggers.TimeTrigger = append(t.Triggers.TimeTrigger, timeTrigger)
}

func (t *Task) addCalendarTrigger(trigger CalendarTrigger) {
	if t.Triggers.CalendarTrigger == nil {
		t.Triggers.CalendarTrigger = []CalendarTrigger{trigger}
		return
	}
	t.Triggers.CalendarTrigger = append(t.Triggers.CalendarTrigger, trigger)
}

func (t *Task) addDailyTrigger(schedule *calendar.Event) {
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a daily trigger", schedule.String())
		return
	}
	// Is it only once a day?
	if len(recurrences) == 1 {
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: recurrences[0].Format(dateFormat),
			ScheduleByDay: &ScheduleByDay{
				DaysInterval: 1,
			},
		})
		return
	}
	// now calculate the difference in between each, and check if they're all the same
	_, compactDifferences := compileDifferences(recurrences)

	if len(compactDifferences) == 1 {
		// case with regular repetition
		interval, _ := period.NewOf(compactDifferences[0])
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: start.Format(dateFormat),
			ScheduleByDay: &ScheduleByDay{
				DaysInterval: 1,
			},
			Repetition: &RepetitionPattern{
				Duration: getRepetionDuration(start, recurrences).Normalise(false),
				Interval: interval.Normalise(false),
			},
		})
		return
	}

	if len(recurrences) > maxTriggers {
		clog.Warningf("this task would need more than %d triggers (%d in total), please rethink your triggers definition", maxTriggers, len(recurrences))
		return
	}
	// install them all
	for _, recurrence := range recurrences {
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: recurrence.Format(dateFormat),
			ScheduleByDay: &ScheduleByDay{
				DaysInterval: 1,
			},
		})
	}
}

func (t *Task) addWeeklyTrigger(schedule *calendar.Event) {
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a weekly trigger", schedule.String())
		return
	}
	// Is it only once per 24h?
	if len(recurrences) == 1 {
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: recurrences[0].Format(dateFormat),
			ScheduleByWeek: &ScheduleByWeek{
				WeeksInterval: 1,
				DaysOfWeek:    convertWeekdays(schedule.WeekDay.GetRangeValues()),
			},
		})
		return
	}
	// now calculate the difference in between each, and check if they're all the same
	_, compactDifferences := compileDifferences(recurrences)

	if len(compactDifferences) == 1 {
		// case with regular repetition
		interval, _ := period.NewOf(compactDifferences[0])
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: start.Format(dateFormat),
			ScheduleByWeek: &ScheduleByWeek{
				WeeksInterval: 1,
				DaysOfWeek:    convertWeekdays(schedule.WeekDay.GetRangeValues()),
			},
			Repetition: &RepetitionPattern{
				Duration: getRepetionDuration(start, recurrences).Normalise(false),
				Interval: interval.Normalise(false),
			},
		})
		return
	}

	if len(recurrences) > maxTriggers {
		clog.Warningf("this task would need more than %d triggers (%d in total), please rethink your triggers definition", maxTriggers, len(recurrences))
		return
	}
	// install them all
	for _, recurrence := range recurrences {
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: recurrence.Format(dateFormat),
			ScheduleByWeek: &ScheduleByWeek{
				WeeksInterval: 1,
				DaysOfWeek:    convertWeekdays(schedule.WeekDay.GetRangeValues()),
			},
		})
	}
}

func convertWeekdays(input []int) DaysOfWeek {
	var weekDays DaysOfWeek
	if len(input) == 0 {
		return weekDays
	}
	for _, weekday := range input {
		switch weekday {
		case 0, 7:
			weekDays.Sunday = WeekDay
		case 1:
			weekDays.Monday = WeekDay
		case 2:
			weekDays.Tuesday = WeekDay
		case 3:
			weekDays.Wednesday = WeekDay
		case 4:
			weekDays.Thursday = WeekDay
		case 5:
			weekDays.Friday = WeekDay
		case 6:
			weekDays.Saturday = WeekDay
		}
	}
	return weekDays
}

func (t *Task) addMonthlyTrigger(schedule *calendar.Event) {
	start := schedule.Next(time.Now())
	// get all recurrences in the same day
	recurrences := schedule.GetAllInBetween(start, start.Add(24*time.Hour))
	if len(recurrences) == 0 {
		clog.Warningf("cannot convert schedule '%s' into a monthly trigger", schedule.String())
		return
	}

	if len(recurrences) > maxTriggers {
		clog.Warningf("this task would need more than %d triggers (%d in total), please rethink your triggers definition", maxTriggers, len(recurrences))
		return
	}
	// install them all
	for _, recurrence := range recurrences {
		if schedule.WeekDay.HasValue() && schedule.Day.HasValue() {
			clog.Warningf("task scheduler does not support a day of the month and a day of the week in the same trigger: %s", schedule.String())
			return
		}
		if schedule.WeekDay.HasValue() {
			t.addCalendarTrigger(CalendarTrigger{
				StartBoundary: recurrence.Format(dateFormat),
				ScheduleByMonthDayOfWeek: &ScheduleByMonthDayOfWeek{
					DaysOfWeek: convertWeekdays(schedule.WeekDay.GetRangeValues()),
					Weeks:      AllWeeks,
					Months:     convertMonths(schedule.Month.GetRangeValues()),
				},
			})
			continue
		}
		t.addCalendarTrigger(CalendarTrigger{
			StartBoundary: recurrence.Format(dateFormat),
			ScheduleByMonth: &ScheduleByMonth{
				DaysOfMonth: convertDaysOfMonth(schedule.Day.GetRangeValues()),
				Months:      convertMonths(schedule.Month.GetRangeValues()),
			},
		})
	}
}

func convertMonths(input []int) Months {
	if len(input) == 0 {
		return Months{
			January:   Month,
			February:  Month,
			March:     Month,
			April:     Month,
			May:       Month,
			June:      Month,
			July:      Month,
			August:    Month,
			September: Month,
			October:   Month,
			November:  Month,
			December:  Month,
		}
	}
	var months Months
	for _, month := range input {
		switch month {
		case 1:
			months.January = Month
		case 2:
			months.February = Month
		case 3:
			months.March = Month
		case 4:
			months.April = Month
		case 5:
			months.May = Month
		case 6:
			months.June = Month
		case 7:
			months.July = Month
		case 8:
			months.August = Month
		case 9:
			months.September = Month
		case 10:
			months.October = Month
		case 11:
			months.November = Month
		case 12:
			months.December = Month
		}
	}
	return months
}

func convertDaysOfMonth(input []int) DaysOfMonth {
	if len(input) == 0 {
		all := make([]int, 31)
		for i := 1; i <= 31; i++ {
			all[i-1] = i
		}
		return DaysOfMonth{all}
	}
	return DaysOfMonth{input}
}
