package schtasks

type Triggers struct {
	CalendarTrigger []CalendarTrigger `xml:"CalendarTrigger"`
}

type CalendarTrigger struct {
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
}
