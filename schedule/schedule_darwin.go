//+build darwin

package schedule

const (
	UserAgentPath   = "~/Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"
)

type LaunchJob struct {
	Label                 string             `plist:"Label"`
	Program               string             `plist:"Program"`
	ProgramArguments      []string           `plist:"ProgramArguments"`
	EnvironmentVariables  map[string]string  `plist:"EnvironmentVariables,omitempty"`
	StandardInPath        string             `plist:"StandardInPath,omitempty"`
	StandardOutPath       string             `plist:"StandardOutPath,omitempty"`
	StandardErrorPath     string             `plist:"StandardErrorPath,omitempty"`
	WorkingDirectory      string             `plist:"WorkingDirectory"`
	StartInterval         int                `plist:"StartInterval,omitempty"`
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval,omitempty"`
}

type CalendarInterval struct {
	Month   int `plist:"Month,omitempty"`   // Month of year (1..12, 1 being January)
	Day     int `plist:"Day,omitempty"`     // Day of month (1..31)
	Weekday int `plist:"Weekday,omitempty"` // Day of week (0..7, 0 and 7 being Sunday)
	Hour    int `plist:"Hour,omitempty"`    // Hour of day (0..23)
	Minute  int `plist:"Minute,omitempty"`  // Minute of hour (0..59)
}
