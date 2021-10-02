package schedule

import (
	"errors"
	"os/exec"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
)

const (
	displayHeader = "\nAnalyzing %s schedule %d/%d\n=================================\n"
)

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

func displayParsedSchedules(command string, events []*calendar.Event) {
	now := time.Now().Round(time.Second)
	for index, event := range events {
		term.Printf(displayHeader, command, index+1, len(events))
		next := event.Next(now)
		term.Printf("  Original form: %s\n", event.Input())
		term.Printf("Normalized form: %s\n", event.String())
		term.Printf("    Next elapse: %s\n", next.Format(time.UnixDate))
		term.Printf("       (in UTC): %s\n", next.UTC().Format(time.UnixDate))
		term.Printf("       From now: %s left\n", next.Sub(now))
		events = append(events, event)
	}
	term.Print("\n")
}

func displaySystemdSchedules(command string, schedules []string) error {
	for index, schedule := range schedules {
		if schedule == "" {
			return errors.New("empty schedule")
		}
		term.Printf(displayHeader, command, index+1, len(schedules))
		cmd := exec.Command("systemd-analyze", "calendar", schedule)
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	term.Print("\n")
	return nil
}
