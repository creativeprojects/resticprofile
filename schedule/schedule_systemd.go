//+build !darwin,!windows

package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/config"
)

func CreateJob(configFile string, profile *config.Profile) error {
	return nil
}

func RemoveJob(profileName string) error {
	return nil
}

func loadSchedules(schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing schedule %d/%d\n========================\n", index+1, len(schedules))
		cmd := exec.Command("systemd-analyze", "calendar", schedule)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return events, err
		}
	}
	return events, nil
}
