//+build darwin windows

package schedule

import (
	"errors"
	"fmt"
	"time"

	"github.com/creativeprojects/resticprofile/calendar"
)

func loadSchedules(schedules []string) ([]*calendar.Event, error) {
	now := time.Now().Round(time.Second)
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing schedule %d/%d\n========================\n", index+1, len(schedules))
		event := calendar.NewEvent()
		err := event.Parse(schedule)
		if err != nil {
			return events, err
		}
		next := event.Next(now)
		fmt.Printf("  Original form: %s\n", schedule)
		fmt.Printf("Normalized form: %s\n", event.String())
		fmt.Printf("    Next elapse: %s\n", next.Format(time.UnixDate))
		fmt.Printf("       (in UTC): %s\n", next.UTC().Format(time.UnixDate))
		fmt.Printf("       From now: %s left\n", next.Sub(now))
		events = append(events, event)
	}
	fmt.Print("\n")
	return events, nil
}
