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
	launchdBin      = "launchd"
	launchctlBin    = "launchctl"
	commandStart    = "start"
	commandStop     = "stop"
	commandLoad     = "load"
	commandUnload   = "unload"
	UserAgentPath   = "Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"

	namePrefix     = "local.resticprofile"
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
func (j *Job) createJob(schedules []*calendar.Event) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	name := getJobName(j.config.Title(), j.config.SubTitle())
	job := &LaunchJob{
		Label:                 name,
		Program:               j.config.Command(),
		ProgramArguments:      append([]string{j.config.Command()}, j.config.Arguments()...),
		EnvironmentVariables:  j.config.Environment(),
		StandardOutPath:       name + ".log",
		StandardErrorPath:     name + ".log",
		WorkingDirectory:      j.config.WorkingDirectory(),
		StartCalendarInterval: getCalendarIntervalsFromSchedules(schedules),
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

	// load the service
	filename := path.Join(home, UserAgentPath, name+agentExtension)
	cmd := exec.Command(launchctlBin, commandLoad, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	// start the service
	cmd = exec.Command(launchctlBin, commandStart, name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// removeJob stops and unloads the agent from launchd, then removes the configuration file
func (j *Job) removeJob() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	name := getJobName(j.config.Title(), j.config.SubTitle())
	filename := path.Join(home, UserAgentPath, name+agentExtension)

	if _, err := os.Stat(filename); err == nil || os.IsExist(err) {
		// stop the service
		stop := exec.Command(launchctlBin, commandStop, name)
		stop.Stdout = os.Stdout
		stop.Stderr = os.Stderr
		// keep going if there's an error here
		_ = stop.Run()

		// unload the service
		unload := exec.Command(launchctlBin, commandUnload, filename)
		unload.Stdout = os.Stdout
		unload.Stderr = os.Stderr
		err = unload.Run()
		if err != nil {
			return err
		}
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	return nil
}

// checkSystem verifies launchd is available on this system
func checkSystem() error {
	found, err := exec.LookPath(launchdBin)
	if err != nil || found == "" {
		return errors.New("it doesn't look like launchd is installed on your system")
	}
	return nil
}

func (j *Job) displayStatus(command string) error {
	cmd := exec.Command(launchctlBin, "list", getJobName(j.config.Title(), j.config.SubTitle()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

func getJobName(profileName, command string) string {
	return fmt.Sprintf("%s.%s.%s", namePrefix, strings.ToLower(profileName), command)
}

// getCalendarIntervalsFromSchedules converts schedules into launchd calendar events
// let's say we've setup these rules:
// Mon-Fri *-*-* *:0,30:00  = every half hour
// Sat     *-*-* 0,12:00:00 = twice a day on saturday
//         *-*-01 *:*:*     = the first of each month
//
// it should translate as:
// 1st rule
//    Weekday = Monday, Minute = 0
//    Weekday = Monday, Minute = 30
//    ... same from Tuesday to Thurday
//    Weekday = Friday, Minute = 0
//    Weekday = Friday, Minute = 30
// Total of 10 rules
// 2nd rule
//    Weekday = Saturday, Hour = 0
//    Weekday = Saturday, Hour = 12
// Total of 2 rules
// 3rd rule
//    Day = 1
// Total of 1 rule
func getCalendarIntervalsFromSchedules(schedules []*calendar.Event) []CalendarInterval {
	entries := make([]CalendarInterval, 0, len(schedules))
	for _, schedule := range schedules {
		entries = append(entries, getCalendarIntervalsFromSchedule(schedule)...)
	}
	return entries
}

func getCalendarIntervalsFromSchedule(schedule *calendar.Event) []CalendarInterval {
	fields := []*calendar.Value{
		schedule.WeekDay,
		schedule.Month,
		schedule.Day,
		schedule.Hour,
		schedule.Minute,
	}

	// create list of permutable items
	total, items := getCombinationItemsFromCalendarValues(fields)

	generateCombination(items, total)

	entries := make([]CalendarInterval, total)

	return entries
}

func permutations(total, num int) int {
	if total == 0 {
		return num
	}
	return total * num
}

func getCombinationItemsFromCalendarValues(fields []*calendar.Value) (int, []combinationItem) {
	// how many entries will I need?
	total := 0
	// list of items for the permutation
	items := []combinationItem{}
	// create list of permutable items
	for _, field := range fields {
		if field.HasValue() {
			values := field.GetRangeValues()
			num := len(values)
			total = permutations(total, num)
			for _, value := range values {
				items = append(items, combinationItem{
					itemType: field.GetType(),
					value:    value,
				})
			}
		}
	}
	return total, items
}
