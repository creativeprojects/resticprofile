//+build darwin

// Documentation about launchd plist file format:
// https://www.launchd.info

package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/creativeprojects/resticprofile/calendar"
	"howett.net/plist"
)

// Default paths for launchd files
const (
	UserAgentPath   = "Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"

	namePrefix     = "local.resticprofile."
	agentExtension = ".agent.plist"
)

// LaunchJob is an agent definition for launchd
type LaunchJob struct {
	Label                 string             `plist:"Label"`
	Program               string             `plist:"Program"`
	ProgramArguments      []string           `plist:"ProgramArguments"`
	EnvironmentVariables  map[string]string  `plist:"EnvironmentVariables,omitempty"`
	StandardInPath        string             `plist:"StandardInPath,omitempty"`
	StandardOutPath       string             `plist:"StandardOutPath,omitempty"`
	StandardErrorPath     string             `plist:"StandardErrorPath,omitempty"`
	WorkingDirectory      string             `plist:"WorkingDirectory"`
	StartInterval         int                `plist:"StartInterval,omitempty"`
	StartCalendarInterval []CalendarInterval `plist:"StartCalendarInterval,omitempty"`
}

// CalendarInterval contains date and time trigger definition
type CalendarInterval struct {
	Month   int `plist:"Month,omitempty"`   // Month of year (1..12, 1 being January)
	Day     int `plist:"Day,omitempty"`     // Day of month (1..31)
	Weekday int `plist:"Weekday,omitempty"` // Day of week (0..7, 0 and 7 being Sunday)
	Hour    int `plist:"Hour,omitempty"`    // Hour of day (0..23)
	Minute  int `plist:"Minute,omitempty"`  // Minute of hour (0..59)
}

// createJob creates a plist file and register it with launchd
func (j *Job) createJob() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binary := absolutePathToBinary(wd, os.Args[0])

	name := namePrefix + strings.ToLower(j.profile.Name)
	job := &LaunchJob{
		Label:   name,
		Program: binary,
		ProgramArguments: []string{
			binary,
			"--no-ansi",
			"--config",
			j.configFile,
			"--name",
			j.profile.Name,
			"backup",
		},
		EnvironmentVariables: j.profile.Environment,
		StandardOutPath:      name + ".log",
		StandardErrorPath:    name + ".error.log",
		WorkingDirectory:     wd,
		StartInterval:        300,
	}

	file, err := os.Create(path.Join(home, UserAgentPath, name+agentExtension))
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := plist.NewEncoder(file)
	encoder.Indent("\t")
	err = encoder.Encode(job)
	if err != nil {
		return err
	}

	return nil
}

// RemoveJob stops and unloads the agent from launchd, then removes the configuration file
func RemoveJob(profileName string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	name := namePrefix + strings.ToLower(profileName)

	stop := exec.Command("launchctl", "stop", name)
	_ = stop.Run()

	unload := exec.Command("launchctl", "unload", name)
	_ = unload.Run()

	return os.Remove(path.Join(home, UserAgentPath, name+agentExtension))
}

func (j *Job) displayStatus() error {
	return nil
}

func loadSchedules(schedules []string) ([]*calendar.Event, error) {
	events := make([]*calendar.Event, 0, len(schedules))
	for index, schedule := range schedules {
		if schedule == "" {
			return events, errors.New("empty schedule")
		}
		fmt.Printf("\nAnalyzing schedule %d/%d\n========================\n", index+1, len(schedules))
	}
	return events, nil
}
