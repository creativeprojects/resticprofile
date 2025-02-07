package schtasks

import (
	"encoding/xml"
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
	Principals struct {
		Principal Principal `xml:"Principal"`
	} `xml:"Principals"`
	Settings struct {
		MultipleInstancesPolicy    string       `xml:"MultipleInstancesPolicy"`
		DisallowStartIfOnBatteries string       `xml:"DisallowStartIfOnBatteries"`
		StopIfGoingOnBatteries     string       `xml:"StopIfGoingOnBatteries"`
		AllowHardTerminate         string       `xml:"AllowHardTerminate"`
		StartWhenAvailable         string       `xml:"StartWhenAvailable"`
		RunOnlyIfNetworkAvailable  string       `xml:"RunOnlyIfNetworkAvailable"`
		IdleSettings               IdleSettings `xml:"IdleSettings"`
		AllowStartOnDemand         string       `xml:"AllowStartOnDemand"`
		Enabled                    string       `xml:"Enabled"`
		Hidden                     string       `xml:"Hidden"`
		RunOnlyIfIdle              string       `xml:"RunOnlyIfIdle"`
		WakeToRun                  string       `xml:"WakeToRun"`
		ExecutionTimeLimit         string       `xml:"ExecutionTimeLimit"`
		Priority                   string       `xml:"Priority"`
	} `xml:"Settings"`
	Actions struct {
		Context string     `xml:"Context,attr"`
		Exec    ExecAction `xml:"Exec"`
	} `xml:"Actions"`
}

type RegistrationInfo struct {
	Date        string `xml:"Date"`
	Author      string `xml:"Author"`
	Description string `xml:"Description"`
	URI         string `xml:"URI"`
}

type Principal struct {
	ID        string `xml:"id,attr"`
	UserId    string `xml:"UserId"`
	LogonType string `xml:"LogonType"`
	RunLevel  string `xml:"RunLevel"`
}

type ExecAction struct {
	Command          string `xml:"Command"`
	Arguments        string `xml:"Arguments"`
	WorkingDirectory string `xml:"WorkingDirectory"`
}

// IdleSettings specifies how the Task Scheduler performs tasks when the computer is in an idle condition.
type IdleSettings struct {
	Duration      period.Period `xml:"Duration"`      // the amount of time that the computer must be in an idle state before the task is run
	RestartOnIdle bool          `xml:"RestartOnIdle"` // whether the task is restarted when the computer cycles into an idle condition more than once
	StopOnIdleEnd bool          `xml:"StopOnIdleEnd"` // indicates that the Task Scheduler will terminate the task if the idle condition ends before the task is completed
	WaitTimeout   period.Period `xml:"WaitTimeout"`   // the amount of time that the Task Scheduler will wait for an idle condition to occur
}

func NewTask() Task {
	task := Task{
		Version: "1.2",
		Xmlns:   "http://schemas.microsoft.com/windows/2004/02/mit/task",
		RegistrationInfo: RegistrationInfo{
			Date: time.Now().Format(time.RFC3339),
		},
	}
	return task
}
