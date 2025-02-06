package schtasks

import "encoding/xml"

type Task struct {
	XMLName          xml.Name `xml:"Task"`
	Text             string   `xml:",chardata"`
	Version          string   `xml:"version,attr"`
	Xmlns            string   `xml:"xmlns,attr"`
	RegistrationInfo struct {
		Text        string `xml:",chardata"`
		Date        string `xml:"Date"`
		Author      string `xml:"Author"`
		Description string `xml:"Description"`
		URI         string `xml:"URI"`
	} `xml:"RegistrationInfo"`
	Triggers struct {
		Text            string `xml:",chardata"`
		CalendarTrigger []struct {
			Text       string `xml:",chardata"`
			Repetition struct {
				Text              string `xml:",chardata"`
				Interval          string `xml:"Interval"`
				Duration          string `xml:"Duration"`
				StopAtDurationEnd string `xml:"StopAtDurationEnd"`
			} `xml:"Repetition"`
			StartBoundary      string `xml:"StartBoundary"`
			ExecutionTimeLimit string `xml:"ExecutionTimeLimit"`
			Enabled            string `xml:"Enabled"`
			ScheduleByWeek     struct {
				Text       string `xml:",chardata"`
				DaysOfWeek struct {
					Text      string `xml:",chardata"`
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
				Text  string `xml:",chardata"`
				Weeks struct {
					Text string   `xml:",chardata"`
					Week []string `xml:"Week"`
				} `xml:"Weeks"`
				DaysOfWeek struct {
					Text   string `xml:",chardata"`
					Monday string `xml:"Monday"`
				} `xml:"DaysOfWeek"`
				Months struct {
					Text     string `xml:",chardata"`
					November string `xml:"November"`
					December string `xml:"December"`
				} `xml:"Months"`
			} `xml:"ScheduleByMonthDayOfWeek"`
		} `xml:"CalendarTrigger"`
	} `xml:"Triggers"`
	Principals struct {
		Text      string `xml:",chardata"`
		Principal struct {
			Text      string `xml:",chardata"`
			ID        string `xml:"id,attr"`
			UserId    string `xml:"UserId"`
			LogonType string `xml:"LogonType"`
			RunLevel  string `xml:"RunLevel"`
		} `xml:"Principal"`
	} `xml:"Principals"`
	Settings struct {
		Text                       string `xml:",chardata"`
		MultipleInstancesPolicy    string `xml:"MultipleInstancesPolicy"`
		DisallowStartIfOnBatteries string `xml:"DisallowStartIfOnBatteries"`
		StopIfGoingOnBatteries     string `xml:"StopIfGoingOnBatteries"`
		AllowHardTerminate         string `xml:"AllowHardTerminate"`
		StartWhenAvailable         string `xml:"StartWhenAvailable"`
		RunOnlyIfNetworkAvailable  string `xml:"RunOnlyIfNetworkAvailable"`
		IdleSettings               struct {
			Text          string `xml:",chardata"`
			Duration      string `xml:"Duration"`
			WaitTimeout   string `xml:"WaitTimeout"`
			StopOnIdleEnd string `xml:"StopOnIdleEnd"`
			RestartOnIdle string `xml:"RestartOnIdle"`
		} `xml:"IdleSettings"`
		AllowStartOnDemand string `xml:"AllowStartOnDemand"`
		Enabled            string `xml:"Enabled"`
		Hidden             string `xml:"Hidden"`
		RunOnlyIfIdle      string `xml:"RunOnlyIfIdle"`
		WakeToRun          string `xml:"WakeToRun"`
		ExecutionTimeLimit string `xml:"ExecutionTimeLimit"`
		Priority           string `xml:"Priority"`
	} `xml:"Settings"`
	Actions struct {
		Text    string `xml:",chardata"`
		Context string `xml:"Context,attr"`
		Exec    struct {
			Text             string `xml:",chardata"`
			Command          string `xml:"Command"`
			Arguments        string `xml:"Arguments"`
			WorkingDirectory string `xml:"WorkingDirectory"`
		} `xml:"Exec"`
	} `xml:"Actions"`
}
