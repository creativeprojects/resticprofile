package crond

import (
	"fmt"
	"io"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
)

// Entry represents a new line in the crontab
type Entry struct {
	event       *calendar.Event
	configFile  string
	profileName string
	commandName string
	commandLine string
	workDir     string
	user        string
}

// NewEntry creates a new crontab entry
func NewEntry(event *calendar.Event, configFile, profileName, commandName, commandLine, workDir string) Entry {
	return Entry{
		event:       event,
		configFile:  configFile,
		profileName: profileName,
		commandName: commandName,
		commandLine: commandLine,
		workDir:     workDir,
	}
}

// WithUser creates a new entry that adds a user that should run the command
func (e Entry) WithUser(user string) Entry {
	e.user = strings.TrimSpace(user)
	return e
}

func (e Entry) HasUser() bool { return len(e.user) > 0 }

func (e Entry) NeedsUser() bool { return e.user == "*" }

func (e Entry) SkipUser() bool { return e.NeedsUser() || e.user == "-" }

// String returns the crontab line representation of the entry (end of line included)
func (e Entry) String() string {
	minute, hour, dayOfMonth, month, dayOfWeek := "*", "*", "*", "*", "*"
	wd := ""
	if e.workDir != "" {
		wd = fmt.Sprintf("cd %s && ", e.workDir)
	}
	if e.event.Minute.HasValue() {
		minute = formatRange(e.event.Minute.GetRanges(), twoDecimals)
	}
	if e.event.Hour.HasValue() {
		hour = formatRange(e.event.Hour.GetRanges(), twoDecimals)
	}
	if e.event.Day.HasValue() {
		dayOfMonth = formatRange(e.event.Day.GetRanges(), twoDecimals)
	}
	if e.event.Month.HasValue() {
		month = formatRange(e.event.Month.GetRanges(), twoDecimals)
	}
	if e.event.WeekDay.HasValue() {
		// don't make ranges for days of the week as it can fail with high sunday (7)
		dayOfWeek = formatList(e.event.WeekDay.GetRangeValues(), formatWeekDay)
	}
	if e.HasUser() && !e.SkipUser() {
		return fmt.Sprintf("%s %s %s %s %s\t%s\t%s%s\n", minute, hour, dayOfMonth, month, dayOfWeek, e.user, wd, e.commandLine)
	}
	return fmt.Sprintf("%s %s %s %s %s\t%s%s\n", minute, hour, dayOfMonth, month, dayOfWeek, wd, e.commandLine)
}

// Generate writes a cron line in the StringWriter (end of line included)
func (e Entry) Generate(w io.StringWriter) error {
	_, err := w.WriteString(e.String())
	return err
}

func (e Entry) Event() *calendar.Event {
	return e.event
}
func (e Entry) ConfigFile() string {
	return e.configFile
}
func (e Entry) ProfileName() string {
	return e.profileName
}
func (e Entry) CommandName() string {
	return e.commandName
}
func (e Entry) CommandLine() string {
	return e.commandLine
}
func (e Entry) WorkDir() string {
	return e.workDir
}
func (e Entry) User() string {
	return e.user
}

func formatWeekDay(weekDay int) string {
	if weekDay >= 7 {
		weekDay -= 7
	}
	return fmt.Sprintf("%d", weekDay)
}

func twoDecimals(value int) string {
	return fmt.Sprintf("%02d", value)
}

func formatList(values []int, formatter func(int) string) string {
	output := make([]string, len(values))
	for i, value := range values {
		output[i] = formatter(value)
	}
	return strings.Join(output, ",")
}

func formatRange(values []calendar.Range, formatter func(int) string) string {
	output := make([]string, len(values))
	for i, value := range values {
		if value.End-value.Start > 1 || value.Start-value.End > 1 {
			// proper range
			output[i] = formatter(value.Start) + "-" + formatter(value.End)
		} else if value.End != value.Start {
			// contiguous values
			output[i] = formatter(value.Start) + "," + formatter(value.End)
		} else {
			// single value
			output[i] = formatter(value.Start)
		}
	}
	return strings.Join(output, ",")
}
