//+build !darwin,!windows

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
}

// NewEntry creates a new crontab entry
func NewEntry(event *calendar.Event, configFile, profileName, commandName, commandLine string) Entry {
	return Entry{
		event:       event,
		configFile:  configFile,
		profileName: profileName,
		commandName: commandName,
		commandLine: commandLine,
	}
}

// String returns the crontab line representation of the entry (end of line included)
func (e Entry) String() string {
	minute, hour, dayOfMonth, month, dayOfWeek := "*", "*", "*", "*", "*"
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
	return fmt.Sprintf("%s %s %s %s %s\t%s\n", minute, hour, dayOfMonth, month, dayOfWeek, e.commandLine)
}

// Generate writes a cron line in the StringWriter (end of line included)
func (e Entry) Generate(w io.StringWriter) error {
	_, err := w.WriteString(e.String())
	return err
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
