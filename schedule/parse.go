package schedule

//
// Common code for all systems
//

import (
	"errors"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/term"
)

func loadParsedSchedules(command string, schedules []string) ([]*calendar.Event, error) {
	now := time.Now().Round(time.Second)
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		term.Printf("\nAnalyzing %s schedule %d/%d\n=================================\n", command, index+1, len(schedules))
		event := calendar.NewEvent()
		err := event.Parse(schedule)
		if err != nil {
			return events, err
		}
		next := event.Next(now)
		term.Printf("  Original form: %s\n", schedule)
		term.Printf("Normalized form: %s\n", event.String())
		term.Printf("    Next elapse: %s\n", next.Format(time.UnixDate))
		term.Printf("       (in UTC): %s\n", next.UTC().Format(time.UnixDate))
		term.Printf("       From now: %s left\n", next.Sub(now))
		events = append(events, event)
	}
	term.Print("\n")
	return events, nil
}
