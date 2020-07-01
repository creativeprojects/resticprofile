//+build darwin

package schedule

import (
	"os"
	"path"
	"strings"

	"github.com/creativeprojects/resticprofile/config"
	"howett.net/plist"
)

const (
	UserAgentPath   = "Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"
)

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

type CalendarInterval struct {
	Month   int `plist:"Month,omitempty"`   // Month of year (1..12, 1 being January)
	Day     int `plist:"Day,omitempty"`     // Day of month (1..31)
	Weekday int `plist:"Weekday,omitempty"` // Day of week (0..7, 0 and 7 being Sunday)
	Hour    int `plist:"Hour,omitempty"`    // Hour of day (0..23)
	Minute  int `plist:"Minute,omitempty"`  // Minute of hour (0..59)
}

func CreateJob(configFile string, profile *config.Profile) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	binary := absolutePathToBinary(wd, os.Args[0])

	name := "local.resticprofile." + strings.ToLower(profile.Name)
	job := &LaunchJob{
		Label:   name,
		Program: binary,
		ProgramArguments: []string{
			binary,
			"--no-ansi",
			"--config",
			config,
			"--name",
			profile.Name,
			"backup",
		},
		EnvironmentVariables: profile.Environment,
		StandardOutPath:      name + ".log",
		StandardErrorPath:    name + ".error.log",
		WorkingDirectory:     wd,
		StartInterval:        300,
	}

	file, err := os.Create(path.Join(home, UserAgentPath, name+".agent.plist"))
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

func RemoveJob(profileName string) error {
	return nil
}
