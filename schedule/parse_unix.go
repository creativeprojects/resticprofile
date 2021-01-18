//+build !darwin,!windows

package schedule

//
// Common code for systemd and crond only
//

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
)

func loadSchedules(command string, schedules []string) ([]*calendar.Event, error) {
	if Scheduler == constants.SchedulerCrond {
		return loadParsedSchedules(command, schedules)
	}
	return loadSystemdSchedules(command, schedules)
}

func loadSystemdSchedules(command string, schedules []string) ([]*calendar.Event, error) {
	for index, schedule := range schedules {
		if schedule == "" {
			return nil, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing %s schedule %d/%d\n=================================\n", command, index+1, len(schedules))
		cmd := exec.Command("systemd-analyze", "calendar", schedule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
	}
	fmt.Print("\n")
	// systemd won't use the parsed events, we can safely return nil
	return nil, nil
}
