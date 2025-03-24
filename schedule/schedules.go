package schedule

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/term"
)

func displayHeader(profile, command string, index, total int) {
	term.Print(platform.LineSeparator)
	header := fmt.Sprintf("Profile (or Group) %s: %s schedule", profile, command)
	if total > 1 {
		header += fmt.Sprintf(" %d/%d", index, total)
	}
	term.Print(header)
	term.Print(platform.LineSeparator)
	term.Print(strings.Repeat("=", len(header)))
	term.Print(platform.LineSeparator)
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

func displayParsedSchedules(profile, command string, events []*calendar.Event) {
	now := time.Now().Round(time.Second)
	for index, event := range events {
		displayHeader(profile, command, index+1, len(events))
		next := event.Next(now)
		term.Printf("  Original form: %s\n", event.Input())
		term.Printf("Normalized form: %s\n", event.String())
		if next.IsZero() {
			term.Print("    Next elapse: Never\n")
			continue
		}
		term.Printf("    Next elapse: %s\n", next.Format(time.UnixDate))
		term.Printf("       (in UTC): %s\n", next.UTC().Format(time.UnixDate))
		term.Printf("       From now: %s left\n", next.Sub(now))
	}
	term.Print(platform.LineSeparator)
}

func displaySystemdSchedules(profile, command string, schedules []string) error {
	for index, schedule := range schedules {
		if schedule == "" {
			return errors.New("empty schedule")
		}
		displayHeader(profile, command, index+1, len(schedules))
		cmd := exec.Command("/usr/bin/systemd-analyze", "calendar", schedule)
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	term.Print(platform.LineSeparator)
	return nil
}
