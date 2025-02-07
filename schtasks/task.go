package schtasks

import (
	"encoding/xml"
	"os/user"
	"time"

	"github.com/rickb777/date/period"
)

type Task struct {
	XMLName          xml.Name         `xml:"Task"`
	Version          string           `xml:"version,attr"`
	Xmlns            string           `xml:"xmlns,attr"`
	RegistrationInfo RegistrationInfo `xml:"RegistrationInfo"`
	Triggers         struct {
		CalendarTrigger []struct {
			Repetition struct {
				Interval          string `xml:"Interval"`
				Duration          string `xml:"Duration"`
				StopAtDurationEnd string `xml:"StopAtDurationEnd"`
			} `xml:"Repetition"`
			StartBoundary      string `xml:"StartBoundary"`
			ExecutionTimeLimit string `xml:"ExecutionTimeLimit"`
			Enabled            string `xml:"Enabled"`
			ScheduleByWeek     struct {
				DaysOfWeek struct {
					Monday    string `xml:"Monday"`
					Tuesday   string `xml:"Tuesday"`
					Wednesday string `xml:"Wednesday"`
					Thursday  string `xml:"Thursday"`
					Friday    string `xml:"Friday"`
					Sunday    string `xml:"Sunday"`
					Saturday  string `xml:"Saturday"`
				} `xml:"DaysOfWeek"`
				WeeksInterval string `xml:"WeeksInterval"`
			} `xml:"ScheduleByWeek"`
			ScheduleByMonthDayOfWeek struct {
				Weeks struct {
					Week []string `xml:"Week"`
				} `xml:"Weeks"`
				DaysOfWeek struct {
					Monday string `xml:"Monday"`
				} `xml:"DaysOfWeek"`
				Months struct {
					November string `xml:"November"`
					December string `xml:"December"`
				} `xml:"Months"`
			} `xml:"ScheduleByMonthDayOfWeek"`
		} `xml:"CalendarTrigger"`
	} `xml:"Triggers"`
	Principals Principals `xml:"Principals"`
	Settings   Settings   `xml:"Settings"`
	Actions    Actions    `xml:"Actions"`
}

type RegistrationInfo struct {
	Date        string `xml:"Date"`
	Author      string `xml:"Author"`
	Description string `xml:"Description"`
	URI         string `xml:"URI"`
}

func NewTask() Task {
	var userName, userID string
	if currentUser, err := user.Current(); err == nil {
		userID = currentUser.Uid
		userName = currentUser.Username
	}
	task := Task{
		Version: "1.2",
		Xmlns:   "http://schemas.microsoft.com/windows/2004/02/mit/task",
		RegistrationInfo: RegistrationInfo{
			Date:   time.Now().Format(time.RFC3339),
			Author: userName,
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
		},
	}
	return task
}
