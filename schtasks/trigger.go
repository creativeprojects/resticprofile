package schtasks

import (
	"github.com/rickb777/date/period"
)

type Triggers struct {
	CalendarTrigger []CalendarTrigger `xml:"CalendarTrigger"`
	TimeTrigger     []TimeTrigger     `xml:"TimeTrigger"`
}

type TimeTrigger struct {
	Enabled            *bool          `xml:"Enabled"` // indicates whether the trigger is enabled
	StartBoundary      string         `xml:"StartBoundary"`
	ExecutionTimeLimit *period.Period `xml:"ExecutionTimeLimit"`
	RandomDelay        *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
}

type CalendarTrigger struct {
	StartBoundary            string                    `xml:"StartBoundary,omitempty"` // the date and time when the trigger is activated
	EndBoundary              string                    `xml:"EndBoundary,omitempty"`   // the date and time when the trigger is deactivated
	Repetition               *RepetitionPattern        `xml:"Repetition"`
	ExecutionTimeLimit       *period.Period            `xml:"ExecutionTimeLimit"` // the maximum amount of time that the task launched by this trigger is allowed to run
	Enabled                  *bool                     `xml:"Enabled"`            // indicates whether the trigger is enabled
	ScheduleByDay            *ScheduleByDay            `xml:"ScheduleByDay,omitempty"`
	ScheduleByWeek           *ScheduleByWeek           `xml:"ScheduleByWeek,omitempty"`
	ScheduleByMonth          *ScheduleByMonth          `xml:"ScheduleByMonth,omitempty"`
	ScheduleByMonthDayOfWeek *ScheduleByMonthDayOfWeek `xml:"ScheduleByMonthDayOfWeek,omitempty"`
}

// RepetitionPattern defines how often the task is run and how long the repetition pattern is repeated after the task is started.
type RepetitionPattern struct {
	Interval          period.Period `xml:"Interval"`          // the amount of time between each restart of the task. Required if RepetitionDuration is specified. Minimum time is one minute
	Duration          period.Period `xml:"Duration"`          // how long the pattern is repeated
	StopAtDurationEnd *bool         `xml:"StopAtDurationEnd"` // indicates if a running instance of the task is stopped at the end of the repetition pattern duration
}

type ScheduleByDay struct {
	RandomDelay  *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	DaysInterval int            `xml:"DaysInterval"`
}

type ScheduleByWeek struct {
	RandomDelay   *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	WeeksInterval int            `xml:"WeeksInterval"`
	DaysOfWeek    DaysOfWeek     `xml:"DaysOfWeek"`
}

type ScheduleByMonth struct {
	RandomDelay *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	Months      Months         `xml:"Months"`
	DaysOfMonth DaysOfMonth    `xml:"DaysOfMonth"`
}

type ScheduleByMonthDayOfWeek struct {
	RandomDelay *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	Months      Months         `xml:"Months"`
	Weeks       Weeks          `xml:"Weeks"`
	DaysOfWeek  DaysOfWeek     `xml:"DaysOfWeek"`
}

type DaysOfMonth struct {
	Day []int `xml:"Day"`
}

type Weeks struct {
	Week []string `xml:"Week"`
}

type DaysOfWeek struct {
	Monday    *string `xml:"Monday"`
	Tuesday   *string `xml:"Tuesday"`
	Wednesday *string `xml:"Wednesday"`
	Thursday  *string `xml:"Thursday"`
	Friday    *string `xml:"Friday"`
	Sunday    *string `xml:"Sunday"`
	Saturday  *string `xml:"Saturday"`
}

type Months struct {
	January   *string `xml:"January"`
	February  *string `xml:"February"`
	March     *string `xml:"March"`
	April     *string `xml:"April"`
	May       *string `xml:"May"`
	June      *string `xml:"June"`
	July      *string `xml:"July"`
	August    *string `xml:"August"`
	September *string `xml:"September"`
	October   *string `xml:"October"`
	November  *string `xml:"November"`
	December  *string `xml:"December"`
}

var (
	emptyString = ""
	WeekDay     = &emptyString
	Month       = &emptyString
	AllWeeks    = Weeks{[]string{"1", "2", "3", "4", "Last"}}
)
