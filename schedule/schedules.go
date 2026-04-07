package schedule

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/term"
)

func displayHeader(terminal *term.Terminal, profile, command string, index, total int) {
	terminal.Print(platform.LineSeparator)
	header := fmt.Sprintf("Profile (or Group) %s: %s schedule", profile, command)
	if total > 1 {
		header += fmt.Sprintf(" %d/%d", index, total)
	}
	terminal.Print(header)
	terminal.Print(platform.LineSeparator)
	terminal.Print(strings.Repeat("=", len(header)))
	terminal.Print(platform.LineSeparator)
}

// parseSchedules creates a *calendar.Event from a string
func parseSchedules(schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for _, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		event := calendar.NewEvent()
		err := event.Parse(schedule)
		if err != nil {
			return events, err
		}
		events = append(events, event)
	}
	return events, nil
}

func displayParsedSchedules(terminal *term.Terminal, profile, command string, events []*calendar.Event) {
	now := time.Now().Round(time.Second)
	for index, event := range events {
		displayHeader(terminal, profile, command, index+1, len(events))
		next := event.Next(now)
		terminal.Printf("  Original form: %s\n", event.Input())
		terminal.Printf("Normalized form: %s\n", event.String())
		if next.IsZero() {
			terminal.Print("    Next elapse: never\n")
			continue
		}
		terminal.Printf("    Next elapse: %s\n", next.Format(time.UnixDate))
		terminal.Printf("       (in UTC): %s\n", next.UTC().Format(time.UnixDate))
		terminal.Printf("       From now: %s left\n", next.Sub(now))
	}
	terminal.Print(platform.LineSeparator)
}
