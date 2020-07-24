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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
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
	commandList     = "list"
	UserAgentPath   = "Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"

	namePrefix      = "local.resticprofile"
	agentExtension  = ".agent.plist"
	daemonExtension = ".plist"

	codeServiceNotFound = 113
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

// CalendarInterval contains date and time trigger definition inside a map.
// keys of the map should be:
//  "Month"   Month of year (1..12, 1 being January)
// 	"Day"     Day of month (1..31)
// 	"Weekday" Day of week (0..7, 0 and 7 being Sunday)
// 	"Hour"    Hour of day (0..23)
// 	"Minute"  Minute of hour (0..59)
type CalendarInterval map[string]int

// newCalendarInterval creates a new map of 5 elements
func newCalendarInterval() *CalendarInterval {
	var value CalendarInterval = make(map[string]int, 5)
	return &value
}

func (c *CalendarInterval) clone() *CalendarInterval {
	clone := newCalendarInterval()
	for key, value := range *c {
		(*clone)[key] = value
	}
	return clone
}

// Init verifies launchd is available on this system
func Init() error {
	found, err := exec.LookPath(launchdBin)
	if err != nil || found == "" {
		return errors.New("it doesn't look like launchd is installed on your system")
	}
	return nil
}

// Close does nothing in systemd
func Close() {
}

// createJob creates a plist file and register it with launchd
func (j *Job) createJob(schedules []*calendar.Event) error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to create a system job: please restart resticprofile as root (with sudo)")
	}
	filename, err := j.createPlistFile(schedules)
	if err != nil {
		if filename != "" {
			os.Remove(filename)
		}
		return err
	}

	j.fixFileOwner(filename)
	if err != nil {
		return err
	}

	// load the service
	cmd := exec.Command(launchctlBin, commandLoad, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) createPlistFile(schedules []*calendar.Event) (string, error) {
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

	filename, err := getFilename(name, j.getSchedulePermission())
	if err != nil {
		return "", err
	}
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := plist.NewEncoder(file)
	encoder.Indent("\t")
	err = encoder.Encode(job)
	if err != nil {
		return filename, err
	}
	return filename, nil
}

// fixFileOwner gives the owner back to the user.
// No error is returned as it's not a big deal if we can't change the file permissions
func (j *Job) fixFileOwner(filename string) {
	if j.getSchedulePermission() == constants.SchedulePermissionSystem || os.Geteuid() != 0 {
		// system permission or user hasn't sudoed
		return
	}
	// this is the case of a launchd agent supposed to be of type user, but created by root
	sudoUID, sudoGID := 0, 0
	var err error
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "SUDO_UID=") {
			temp := strings.TrimPrefix(env, "SUDO_UID=")
			sudoUID, err = strconv.Atoi(temp)
			if err != nil {
				return
			}
		}
		if strings.HasPrefix(env, "SUDO_GID=") {
			temp := strings.TrimPrefix(env, "SUDO_GID=")
			sudoGID, err = strconv.Atoi(temp)
			if err != nil {
				return
			}
		}
	}
	if sudoUID > 0 {
		err = os.Chown(filename, sudoUID, sudoGID)
		if err != nil {
			clog.Warningf("cannot change agent owner: %v", err)
		}
	}
}

// removeJob stops and unloads the agent from launchd, then removes the configuration file
func (j *Job) removeJob() error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to remove a system job: please restart resticprofile as root (with sudo)")
	}
	name := getJobName(j.config.Title(), j.config.SubTitle())
	filename, err := getFilename(name, j.getSchedulePermission())
	if err != nil {
		return err
	}

	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return ErrorServiceNotFound
	}
	// stop the service in case it's already running
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

	return nil
}

func (j *Job) displayStatus(command string) error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed view a system job: please restart resticprofile as root (with sudo)")
	}
	cmd := exec.Command(launchctlBin, commandList, getJobName(j.config.Title(), j.config.SubTitle()))
	output, err := cmd.Output()
	if cmd.ProcessState.ExitCode() == codeServiceNotFound {
		return ErrorServiceNotFound
	}
	if err != nil {
		return err
	}
	status := parseStatus(string(output))
	if len(status) == 0 {
		// output was not parsed, it could mean output format has changed
		fmt.Println(string(output))
	}
	// order keys alphabetically
	keys := make([]string, 0, len(status))
	for key := range status {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		fmt.Fprintf(writer, "%s:\t %s\n", key, status[key])
	}
	writer.Flush()
	fmt.Println("")

	return nil
}

func getJobName(profileName, command string) string {
	return fmt.Sprintf("%s.%s.%s", namePrefix, strings.ToLower(profileName), command)
}

func getFilename(name, permission string) (string, error) {
	if permission == constants.SchedulePermissionSystem {
		return path.Join(GlobalDaemons, name+daemonExtension), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, UserAgentPath, name+agentExtension), nil
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
		entries = append(entries, getCalendarIntervalsFromScheduleTree(generateTreeOfSchedules(schedule))...)
	}
	return entries
}

func getCalendarIntervalsFromScheduleTree(tree []*treeElement) []CalendarInterval {
	entries := make([]CalendarInterval, 0)
	for _, element := range tree {
		// creates a new calendar entry for each tip of the branch
		newEntry := newCalendarInterval()
		fillInValueFromScheduleTreeElement(newEntry, element, &entries)
	}
	return entries
}

func fillInValueFromScheduleTreeElement(currentEntry *CalendarInterval, element *treeElement, entries *[]CalendarInterval) {
	setCalendarIntervalValueFromType(currentEntry, element.value, element.elementType)
	if element.subElements == nil || len(element.subElements) == 0 {
		// end of the line, this entry is finished
		*entries = append(*entries, *currentEntry)
		return
	}
	for _, subElement := range element.subElements {
		// new branch means new calendar entry
		fillInValueFromScheduleTreeElement(currentEntry.clone(), subElement, entries)
	}
}

func setCalendarIntervalValueFromType(entry *CalendarInterval, value int, typeValue calendar.TypeValue) {
	if entry == nil {
		entry = newCalendarInterval()
	}
	switch typeValue {
	case calendar.TypeWeekDay:
		(*entry)["Weekday"] = value
	case calendar.TypeMonth:
		(*entry)["Month"] = value
	case calendar.TypeDay:
		(*entry)["Day"] = value
	case calendar.TypeHour:
		(*entry)["Hour"] = value
	case calendar.TypeMinute:
		(*entry)["Minute"] = value
	}
}

func parseStatus(status string) map[string]string {
	expr := regexp.MustCompile(`^\s*"(\w+)"\s*=\s*(.*);$`)
	lines := strings.Split(status, "\n")
	output := make(map[string]string, len(lines))
	for _, line := range lines {
		match := expr.FindStringSubmatch(line)
		if len(match) == 3 {
			output[match[1]] = strings.Trim(match[2], "\"")
		}
	}
	return output
}
