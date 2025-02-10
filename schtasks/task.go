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
	dateFormat  = time.RFC3339
	maxTriggers = 60
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
		Version: "1.2",
		Xmlns:   "http://schemas.microsoft.com/windows/2004/02/mit/task",
		RegistrationInfo: RegistrationInfo{
			Date:   time.Now().Format(dateFormat),
			Author: constants.ApplicationName,
		},
		Principals: Principals{
			Principal: Principal{
				ID:        "Author",
				UserId:    userID,
				LogonType: LogonTypeInteractiveToken,
				RunLevel:  RunLevelLeastPrivilege,
			},
		},
		Settings: Settings{
			AllowDemandStart:           true,
			AllowHardTerminate:         true,
			Compatibility:              TaskCompatibilityV2,
			DisallowStartIfOnBatteries: true,
			Enabled:                    true,
			IdleSettings: IdleSettings{
				Duration:      period.NewHMS(0, 10, 0), // PT10M
				WaitTimeout:   period.NewHMS(1, 0, 0),  // PT1H
				StopOnIdleEnd: true,
			},
			MultipleInstancesPolicy: MultipleInstancesIgnoreNew,
			Priority:                7,
			StopIfGoingOnBatteries:  true,
			ExecutionTimeLimit:      period.NewHMS(72, 0, 0), // PT72H
		},
		Actions: Actions{
			Context: "Author",
			Exec:    make([]ExecAction, 0, 1), // prepare space for 1 command
		},
		Triggers: Triggers{
			TimeTrigger:     make([]TimeTrigger, 0),
			CalendarTrigger: make([]CalendarTrigger, 0),
		},
	}
	return task
}

// AddExecAction returns the same instance of Task (for chaining)
func (t *Task) AddExecAction(action ExecAction) *Task {
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
		// if schedule.IsWeekly() {
		// 	createWeeklyTrigger(task, schedule)
		// 	continue
		// }
		// if schedule.IsMonthly() {
		// 	createMonthlyTrigger(task, schedule)
		// 	continue
		// }
		clog.Warningf("cannot convert schedule '%s' into a task scheduler equivalent", schedule.String())
	}
}

func (t *Task) addTimeTrigger(triggerOnce time.Time) {
	t.Triggers.TimeTrigger = append(t.Triggers.TimeTrigger, TimeTrigger{
		StartBoundary: triggerOnce.Format(dateFormat),
	})
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

func (t *Task) addCalendarTrigger(trigger CalendarTrigger) {
	t.Triggers.CalendarTrigger = append(t.Triggers.CalendarTrigger, trigger)
}
