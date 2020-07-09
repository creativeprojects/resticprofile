//+build !darwin,!windows

package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
)

func loadSchedules(command string, schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing %s schedule %d/%d\n========================\n", command, index+1, len(schedules))
		cmd := exec.Command("systemd-analyze", "calendar", schedule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return events, err
		}
	}
	fmt.Print("\n")
	return events, nil
}
