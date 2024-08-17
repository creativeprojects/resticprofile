//go:build darwin

package schedule

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/spf13/afero"
	"howett.net/plist"
)

// Documentation about launchd plist file format:
// https://www.launchd.info

// Default paths for launchd files
const (
	launchdBin      = "launchd"
	launchctlBin    = "launchctl"
	launchdStart    = "start"
	launchdStop     = "stop"
	launchdLoad     = "load"
	launchdUnload   = "unload"
	launchdList     = "list"
	UserAgentPath   = "Library/LaunchAgents"
	GlobalAgentPath = "/Library/LaunchAgents"
	GlobalDaemons   = "/Library/LaunchDaemons"

	namePrefix      = "local.resticprofile"
	agentExtension  = ".agent.plist"
	daemonExtension = ".plist"

	codeServiceNotFound = 113
)

// LaunchJob is an agent definition for launchd
type LaunchdJob struct {
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
	ProcessType           string             `plist:"ProcessType"`
	LowPriorityIO         bool               `plist:"LowPriorityIO"`
	Nice                  int                `plist:"Nice"`
}

var priorityValues = map[string]string{
	constants.SchedulePriorityBackground: "Background",
	constants.SchedulePriorityStandard:   "Standard",
}

type HandlerLaunchd struct {
	config SchedulerConfig
	fs     afero.Fs
}

// Init verifies launchd is available on this system
func (h *HandlerLaunchd) Init() error {
	return lookupBinary("launchd", launchdBin)
}

// Close does nothing with launchd
func (h *HandlerLaunchd) Close() {
	// nothing to do
}

func (h *HandlerLaunchd) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

func (h *HandlerLaunchd) DisplayParsedSchedules(command string, events []*calendar.Event) {
	displayParsedSchedules(command, events)
}

// DisplaySchedules does nothing with launchd
func (h *HandlerLaunchd) DisplaySchedules(command string, schedules []string) error {
	return nil
}

// DisplayStatus does nothing with launchd
func (h *HandlerLaunchd) DisplayStatus(profileName string) error {
	return nil
}

// CreateJob creates a plist file and registers it with launchd
func (h *HandlerLaunchd) CreateJob(job *Config, schedules []*calendar.Event, permission string) error {
	filename, err := h.createPlistFile(h.getLaunchdJob(job, schedules), permission)
	if err != nil {
		if filename != "" {
			os.Remove(filename)
		}
		return err
	}

	// load the service
	cmd := exec.Command(launchctlBin, launchdLoad, filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	if _, noStart := job.GetFlag("no-start"); !noStart {
		// ask the user if he wants to start the service now
		name := getJobName(job.ProfileName, job.CommandName)
		message := `
By default, a macOS agent access is restricted. If you leave it to start in the background it's likely to fail.
You have to start it manually the first time to accept the requests for access:

%% %s %s %s

Do you want to start it now?`
		answer := term.AskYesNo(os.Stdin, fmt.Sprintf(message, launchctlBin, launchdStart, name), true)
		if answer {
			// start the service
			cmd := exec.Command(launchctlBin, launchdStart, name)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HandlerLaunchd) getLaunchdJob(job *Config, schedules []*calendar.Event) *LaunchdJob {
	name := getJobName(job.ProfileName, job.CommandName)
	args := job.Arguments
	// we always set the log file in the job settings as a default
	// if changed in the configuration via schedule-log the standard output will be empty anyway
	logfile := name + ".log"

	// Format schedule env, adding PATH if not yet provided by the schedule config
	env := util.NewDefaultEnvironment(job.Environment...)
	if !env.Has("PATH") {
		env.Put("PATH", os.Getenv("PATH"))
	}

	lowPriorityIO := true
	nice := constants.DefaultBackgroundNiceFlag
	if job.GetPriority() == constants.SchedulePriorityStandard {
		lowPriorityIO = false
		nice = constants.DefaultStandardNiceFlag
	}

	launchdJob := &LaunchdJob{
		Label:                 name,
		Program:               job.Command,
		ProgramArguments:      append([]string{job.Command, "--no-prio"}, args...),
		StandardOutPath:       logfile,
		StandardErrorPath:     logfile,
		WorkingDirectory:      job.WorkingDirectory,
		StartCalendarInterval: getCalendarIntervalsFromSchedules(schedules),
		EnvironmentVariables:  env.ValuesAsMap(),
		Nice:                  nice,
		ProcessType:           priorityValues[job.GetPriority()],
		LowPriorityIO:         lowPriorityIO,
	}
	return launchdJob
}

func (h *HandlerLaunchd) createPlistFile(launchdJob *LaunchdJob, permission string) (string, error) {
	filename, err := getFilename(launchdJob.Label, permission)
	if err != nil {
		return "", err
	}
	if permission != constants.SchedulePermissionSystem {
		// in some very recent installations of macOS, the user's LaunchAgents folder may not exist
		_ = h.fs.MkdirAll(path.Dir(filename), 0o700)
	}
	file, err := h.fs.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := plist.NewEncoder(file)
	encoder.Indent("\t")
	err = encoder.Encode(launchdJob)
	if err != nil {
		return filename, err
	}
	return filename, nil
}

// RemoveJob stops and unloads the agent from launchd, then removes the configuration file
func (h *HandlerLaunchd) RemoveJob(job *Config, permission string) error {
	name := getJobName(job.ProfileName, job.CommandName)
	filename, err := getFilename(name, permission)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return ErrServiceNotFound
	}
	// stop the service in case it's already running
	stop := exec.Command(launchctlBin, launchdStop, name)
	stop.Stdout = os.Stdout
	stop.Stderr = os.Stderr
	// keep going if there's an error here
	_ = stop.Run()

	// unload the service
	unload := exec.Command(launchctlBin, launchdUnload, filename)
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

func (h *HandlerLaunchd) DisplayJobStatus(job *Config) error {
	permission := getSchedulePermission(job.Permission)
	ok := checkPermission(permission)
	if !ok {
		return permissionError("view")
	}
	cmd := exec.Command(launchctlBin, launchdList, getJobName(job.ProfileName, job.CommandName))
	output, err := cmd.Output()
	if cmd.ProcessState.ExitCode() == codeServiceNotFound {
		return ErrServiceNotFound
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
	writer := tabwriter.NewWriter(term.GetOutput(), 0, 0, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		fmt.Fprintf(writer, "%s:\t %s\n", key, status[key])
	}
	writer.Flush()
	fmt.Println("")

	return nil
}

var (
	_ Handler = &HandlerLaunchd{}
)

// CalendarInterval contains date and time trigger definition inside a map.
// keys of the map should be:
//
//	"Month"   Month of year (1..12, 1 being January)
//	"Day"     Day of month (1..31)
//	"Weekday" Day of week (0..7, 0 and 7 being Sunday)
//	"Hour"    Hour of day (0..23)
//	"Minute"  Minute of hour (0..59)
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
//
//	Mon-Fri *-*-* *:0,30:00  = every half hour
//	Sat     *-*-* 0,12:00:00 = twice a day on saturday
//	        *-*-01 *:*:*     = the first of each month
//
// it should translate as:
// 1st rule
//
//	Weekday = Monday, Minute = 0
//	Weekday = Monday, Minute = 30
//	... same from Tuesday to Thurday
//	Weekday = Friday, Minute = 0
//	Weekday = Friday, Minute = 30
//
// Total of 10 rules
// 2nd rule
//
//	Weekday = Saturday, Hour = 0
//	Weekday = Saturday, Hour = 12
//
// Total of 2 rules
// 3rd rule
//
//	Day = 1
//
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
	if len(element.subElements) == 0 {
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

// init registers HandlerLaunchd
func init() {
	AddHandlerProvider(func(config SchedulerConfig, _ bool) (hr Handler) {
		if config.Type() == constants.SchedulerLaunchd ||
			config.Type() == constants.SchedulerOSDefault {
			hr = &HandlerLaunchd{
				config: config,
				fs:     afero.NewOsFs(),
			}
		}
		return
	})
}
