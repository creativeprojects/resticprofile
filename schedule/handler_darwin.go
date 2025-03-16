//go:build darwin

package schedule

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/darwin"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/spf13/afero"
	"howett.net/plist"
)

// Documentation about launchd plist file format:
// https://www.launchd.info

// Default paths for launchd files
const (
	launchctlBin     = "/bin/launchctl"
	launchdBootstrap = "bootstrap"
	launchdBootout   = "bootout"
	launchdList      = "list"
	UserAgentPath    = "Library/LaunchAgents"
	GlobalAgentPath  = "/Library/LaunchAgents"
	GlobalDaemons    = "/Library/LaunchDaemons"

	namePrefix      = "local.resticprofile." // namePrefix is the prefix used for all launchd job labels managed by resticprofile
	agentExtension  = ".agent.plist"
	daemonExtension = ".plist"

	codeServiceNotFound = 113
)

type HandlerLaunchd struct {
	config SchedulerConfig
	fs     afero.Fs
}

// Init verifies launchd is available on this system
func (h *HandlerLaunchd) Init() error {
	return lookupBinary("launchd", launchctlBin)
}

// Close does nothing with launchd
func (h *HandlerLaunchd) Close() {
	// nothing to do
}

func (h *HandlerLaunchd) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return parseSchedules(schedules)
}

func (h *HandlerLaunchd) DisplaySchedules(profile, command string, schedules []string) error {
	events, err := parseSchedules(schedules)
	if err != nil {
		return err
	}
	displayParsedSchedules(profile, command, events)
	return nil
}

// DisplayStatus does nothing with launchd
func (h *HandlerLaunchd) DisplayStatus(profileName string) error {
	return nil
}

// CreateJob creates a plist file and registers it with launchd
func (h *HandlerLaunchd) CreateJob(job *Config, schedules []*calendar.Event, permission Permission) error {
	filename, err := h.createPlistFile(h.getLaunchdJob(job, schedules), permission)
	if err != nil {
		if filename != "" {
			os.Remove(filename)
		}
		return err
	}

	cmd := launchctlCommand(launchdBootstrap, domainTarget(permission), filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// RemoveJob stops and unloads the agent from launchd, then removes the configuration file
func (h *HandlerLaunchd) RemoveJob(job *Config, permission Permission) error {
	name := getJobName(job.ProfileName, job.CommandName)
	filename, err := getFilename(name, permission)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return ErrScheduledJobNotFound
	}

	unload := launchctlCommand(launchdBootout, domainTarget(permission)+"/"+name)
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
	permission, _ := h.DetectSchedulePermission(PermissionFromConfig(job.Permission))
	if ok := permission.Check(); !ok {
		return permissionError("view")
	}
	cmd := launchctlCommand(launchdList, getJobName(job.ProfileName, job.CommandName))
	output, err := cmd.Output()
	if cmd.ProcessState.ExitCode() == codeServiceNotFound {
		return ErrScheduledJobNotFound
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
		// if slices.Contains([]string{"LimitLoadToSessionType", "OnDemand"}, key) {
		// 	continue
		// }
		keys = append(keys, key)
	}
	sort.Strings(keys)
	writer := tabwriter.NewWriter(term.GetOutput(), 0, 0, 0, ' ', tabwriter.AlignRight)
	for _, key := range keys {
		fmt.Fprintf(writer, "%s:\t %s\n", spacedTitle(key), status[key])
	}
	writer.Flush()
	fmt.Println("")

	return nil
}

func (h *HandlerLaunchd) Scheduled(profileName string) ([]Config, error) {
	jobs := make([]Config, 0)
	if profileName == "" {
		profileName = "*"
	} else {
		profileName = strings.ToLower(profileName)
	}
	// system jobs
	systemJobs := h.getScheduledJob(profileName, constants.SchedulePermissionSystem)
	jobs = append(jobs, systemJobs...)
	// user jobs
	userJobs := h.getScheduledJob(profileName, constants.SchedulePermissionUser)
	jobs = append(jobs, userJobs...)
	return jobs, nil
}

func (h *HandlerLaunchd) getLaunchdJob(job *Config, schedules []*calendar.Event) *darwin.LaunchdJob {
	name := getJobName(job.ProfileName, job.CommandName)
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

	launchdJob := &darwin.LaunchdJob{
		Label:                   name,
		Program:                 job.Command,
		ProgramArguments:        append([]string{job.Command, "--no-prio"}, job.Arguments.RawArgs()...),
		StandardOutPath:         logfile,
		StandardErrorPath:       logfile,
		WorkingDirectory:        job.WorkingDirectory,
		StartCalendarInterval:   darwin.GetCalendarIntervalsFromSchedules(schedules),
		EnvironmentVariables:    env.ValuesAsMap(),
		Nice:                    nice,
		ProcessType:             darwin.NewProcessType(job.GetPriority()),
		LowPriorityIO:           lowPriorityIO,
		LowPriorityBackgroundIO: lowPriorityIO,
		LimitLoadToSessionType:  darwin.NewSessionType(job.Permission),
	}
	return launchdJob
}

func (h *HandlerLaunchd) getScheduledJob(profileName, permission string) []Config {
	matches, err := afero.Glob(h.fs, getSchedulePattern(profileName, permission))
	if err != nil {
		clog.Warningf("Error while listing %s jobs: %s", permission, err)
	}
	jobs := make([]Config, 0, len(matches))
	for _, match := range matches {
		job, err := h.getJobConfig(match)
		if err != nil {
			clog.Warning(err)
			continue
		}
		if job != nil {
			if job.Permission == "" {
				job.Permission = permission
			}
			jobs = append(jobs, *job)
		}
	}
	return jobs
}

// DetectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func (h *HandlerLaunchd) DetectSchedulePermission(permission Permission) (Permission, bool) {
	switch permission {
	case PermissionSystem, PermissionUserBackground, PermissionUserLoggedOn:
		// well defined
		return permission, true

	default:
		// best guess is depending on the user being root or not:
		detected := PermissionUserLoggedOn // sane default
		if os.Geteuid() == 0 {
			detected = PermissionSystem
		}
		// darwin can backup protected files without the need of a system task
		return detected, true
	}
}

func getSchedulePattern(profileName, permission string) string {
	pattern := "%s%s.*%s"
	if permission == constants.SchedulePermissionSystem {
		return fmt.Sprintf(pattern, path.Join(GlobalDaemons, namePrefix), profileName, daemonExtension)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return fmt.Sprintf(pattern, path.Join(home, UserAgentPath, namePrefix), profileName, agentExtension)
}

func getCommandAndProfileFromFilename(filename string) (command string, profile string) {
	// try removing both daemon and agent extensions
	filename = strings.TrimSuffix(filename, agentExtension)  // longer one
	filename = strings.TrimSuffix(filename, daemonExtension) // shorter one
	filename = strings.TrimPrefix(path.Base(filename), namePrefix)
	parts := strings.Split(filename, ".")
	command = parts[len(parts)-1]
	profile = strings.Join(parts[:len(parts)-1], ".")
	return
}

func (h *HandlerLaunchd) getJobConfig(filename string) (*Config, error) {
	commandName, profileName := getCommandAndProfileFromFilename(filename)

	launchdJob, err := h.readPlistFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading plist file: %w", err)
	}

	permission := ""
	switch launchdJob.LimitLoadToSessionType {
	case darwin.SessionTypeDefault, darwin.SessionTypeGUI:
		permission = constants.SchedulePermissionUserLoggedOn
	case darwin.SessionTypeBackground:
		permission = constants.SchedulePermissionUser
	case darwin.SessionTypeSystem:
		permission = constants.SchedulePermissionSystem
	}

	args := NewCommandArguments(launchdJob.ProgramArguments[2:]) // first is binary, second is --no-prio
	job := &Config{
		ProfileName:      profileName,
		CommandName:      commandName,
		Command:          launchdJob.Program,
		ConfigFile:       args.ConfigFile(),
		Arguments:        args,
		WorkingDirectory: launchdJob.WorkingDirectory,
		Schedules:        darwin.ParseCalendarIntervals(launchdJob.StartCalendarInterval),
		Permission:       permission,
	}
	return job, nil
}

func (h *HandlerLaunchd) createPlistFile(launchdJob *darwin.LaunchdJob, permission Permission) (string, error) {
	filename, err := getFilename(launchdJob.Label, permission)
	if err != nil {
		return "", err
	}
	if permission != PermissionSystem {
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

func (h *HandlerLaunchd) readPlistFile(filename string) (*darwin.LaunchdJob, error) {
	file, err := h.fs.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := plist.NewDecoder(file)
	launchdJob := new(darwin.LaunchdJob)
	err = decoder.Decode(launchdJob)
	if err != nil {
		return nil, err
	}
	return launchdJob, nil
}

var (
	_ Handler = &HandlerLaunchd{}
)

func getJobName(profileName, command string) string {
	return fmt.Sprintf("%s%s.%s", namePrefix, strings.ToLower(profileName), command)
}

func getFilename(name string, permission Permission) (string, error) {
	if permission == PermissionSystem {
		return path.Join(GlobalDaemons, name+daemonExtension), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, UserAgentPath, name+agentExtension), nil
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

func domainTarget(permission Permission) string {
	// after a sudo, macOs returns the root user on both os.Getuid() and os.Geteuid()
	// to detect the logged on user after a sudo, we need to use the environment variable
	uid := os.Getuid()
	if userid, sudo := os.LookupEnv("SUDO_UID"); sudo {
		if temp, err := strconv.Atoi(userid); err == nil {
			uid = temp
		}
	}
	switch permission {
	case PermissionSystem:
		return "system"
	case PermissionUserLoggedOn:
		return fmt.Sprintf("gui/%d", uid)

	case PermissionUserBackground:
		return fmt.Sprintf("user/%d", uid)

	default:
		return ""
	}
}

func launchctlCommand(arg ...string) *exec.Cmd {
	clog.Debugf("running command: '%s %s'", launchctlBin, strings.Join(arg, " "))
	return exec.Command(launchctlBin, arg...)
}

type keyValue struct {
	key   string
	value string
}

func parsePrint(output []byte) ([]keyValue, error) {
	keys := []string{"state"}
	info := make([]keyValue, 0, 10)
	lines := bytes.Split(output, []byte{'\n'})
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if key, value, found := bytes.Cut(line, []byte{'='}); found {
			strKey := string(bytes.TrimSpace(key))
			strValue := string(bytes.TrimSpace(value))
			if slices.Contains(keys, strKey) {
				info = append(info, keyValue{strKey, strValue})
			}
		}
	}
	return info, nil
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
