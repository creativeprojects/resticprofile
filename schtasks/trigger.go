package schtasks

import "github.com/rickb777/date/period"

type Triggers struct {
	CalendarTrigger []CalendarTrigger `xml:"CalendarTrigger"`
	TimeTrigger     []TimeTrigger     `xml:"TimeTrigger"`
}

type TimeTrigger struct {
	StartBoundary      string `xml:"StartBoundary"`
	ExecutionTimeLimit string `xml:"ExecutionTimeLimit"`
}

type CalendarTrigger struct {
	Repetition               RepetitionPattern         `xml:"Repetition"`
	StartBoundary            string                    `xml:"StartBoundary,omitempty"` // the date and time when the trigger is activated
	EndBoundary              string                    `xml:"EndBoundary,omitempty"`   // the date and time when the trigger is deactivated
	ExecutionTimeLimit       period.Period             `xml:"ExecutionTimeLimit"`      // the maximum amount of time that the task launched by this trigger is allowed to run
	Enabled                  bool                      `xml:"Enabled"`                 // indicates whether the trigger is enabled
	ScheduleByDay            *ScheduleByDay            `xml:"ScheduleByDay,omitempty"`
	ScheduleByWeek           *ScheduleByWeek           `xml:"ScheduleByWeek,omitempty"`
	ScheduleByMonthDayOfWeek *ScheduleByMonthDayOfWeek `xml:"ScheduleByMonthDayOfWeek,omitempty"`
}

// RepetitionPattern defines how often the task is run and how long the repetition pattern is repeated after the task is started.
type RepetitionPattern struct {
	Duration          period.Period `xml:"Duration"`          // how long the pattern is repeated
	Interval          period.Period `xml:"Interval"`          // the amount of time between each restart of the task. Required if RepetitionDuration is specified. Minimum time is one minute
	StopAtDurationEnd bool          `xml:"StopAtDurationEnd"` // indicates if a running instance of the task is stopped at the end of the repetition pattern duration
}

type ScheduleByDay struct {
	RandomDelay  *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	DaysInterval int            `xml:"DaysInterval"`
}

type ScheduleByWeek struct {
	RandomDelay *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	DaysOfWeek  struct {
		Monday    string `xml:"Monday"`
		Tuesday   string `xml:"Tuesday"`
		Wednesday string `xml:"Wednesday"`
		Thursday  string `xml:"Thursday"`
		Friday    string `xml:"Friday"`
		Sunday    string `xml:"Sunday"`
		Saturday  string `xml:"Saturday"`
	} `xml:"DaysOfWeek"`
	WeeksInterval string `xml:"WeeksInterval"`
}

type ScheduleByMonthDayOfWeek struct {
	RandomDelay *period.Period `xml:"RandomDelay,omitempty"` // a delay time that is randomly added to the start time of the trigger
	Weeks       struct {
		Week []string `xml:"Week"`
	} `xml:"Weeks"`
	DaysOfWeek struct {
		Monday string `xml:"Monday"`
	} `xml:"DaysOfWeek"`
	Months struct {
		November string `xml:"November"`
		December string `xml:"December"`
	} `xml:"Months"`
}
