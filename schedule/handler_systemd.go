//go:build !darwin && !windows

package schedule

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/user"
)

const (
	systemctlStart   = "start"
	systemctlStop    = "stop"
	systemctlEnable  = "enable"
	systemctlDisable = "disable"
	systemctlStatus  = "status"
	systemctlReload  = "daemon-reload"
	flagUserUnit     = "--user"
	flagNoPager      = "--no-pager"
	unitNotFound     = "not-found"

	// https://www.freedesktop.org/software/systemd/man/systemctl.html#Exit%20status
	codeStatusNotRunning   = 3
	codeStatusUnitNotFound = 4
	codeStopUnitNotFound   = 5 // undocumented
)

var (
	journalctlBinary = "journalctl"
	systemctlBinary  = "systemctl"
	analyzeBinary    = "systemd-analyze"
)

// HandlerSystemd is a handler to schedule tasks using systemd
type HandlerSystemd struct {
	config SchedulerSystemd
}

// NewHandlerSystemd creates a new handler to schedule jobs using systemd
func NewHandlerSystemd(config SchedulerConfig) *HandlerSystemd {
	cfg, ok := config.(SchedulerSystemd)
	if !ok {
		cfg = SchedulerSystemd{} // empty configuration
	}
	return &HandlerSystemd{config: cfg}
}

// Init verifies systemd is available on this system
func (h *HandlerSystemd) Init() error {
	return lookupBinary("systemd", systemctlBinary)
}

// Close does nothing with systemd
func (h *HandlerSystemd) Close() {
	// nothing to do
}

// ParseSchedules always returns nil on systemd
func (h *HandlerSystemd) ParseSchedules(schedules []string) ([]*calendar.Event, error) {
	return nil, nil
}

// DisplaySchedules displays the schedules through the systemd-analyze command
func (h *HandlerSystemd) DisplaySchedules(profile, command string, schedules []string) error {
	return displaySystemdSchedules(profile, command, schedules)
}

// Timers summary
// ===============
// NEXT                        LEFT       LAST                        PASSED  UNIT                                  ACTIVATES
// Tue 2024-10-29 18:45:00 GMT 28min left Tue 2024-10-29 18:16:09 GMT 38s ago resticprofile-copy@profile-self.timer resticprofile-copy@profile-self.service
//
// 1 timers listed.
func (h *HandlerSystemd) DisplayStatus(profileName string) error {
	var (
		status string
		err    error
	)
	if os.Geteuid() == 0 {
		// if the user is root, we search for system timers
		status, err = getSystemdStatus(profileName, systemd.SystemUnit)
	} else {
		// otherwise user timers
		status, err = getSystemdStatus(profileName, systemd.UserUnit)
	}
	if err != nil || status == "" || strings.Contains(status, "0 timers") {
		// fail silently
		return nil
	}
	fmt.Fprintf(term.GetOutput(), "\nTimers summary\n===============\n%s\n", status)
	return nil
}

// CreateJob is creating the systemd unit and activating it
func (h *HandlerSystemd) CreateJob(job *Config, schedules []*calendar.Event, permission Permission) error {
	u := user.Current()
	unitType, user := permissionToSystemd(u, permission)

	if unitType == systemd.UserUnit && job.AfterNetworkOnline {
		return fmt.Errorf("after-network-online is not available for \"user_logged_on\" permission schedules")
	}

	timerFile := systemd.GetTimerFile(job.ProfileName, job.CommandName)

	// check the user hasn't changed the permission, which could duplicate the unit (system & user)
	otherUnitType := systemd.SystemUnit
	if unitType == systemd.SystemUnit {
		otherUnitType = systemd.UserUnit
	}
	if existingConfigs, _ := getConfigs(job.ProfileName, otherUnitType); len(existingConfigs) > 0 { // ignore errors here
		for _, cfg := range existingConfigs {
			if cfg.CommandName == job.CommandName && cfg.ProfileName == job.ProfileName {
				// we'd better remove this schedule first
				clog.Infof("removing existing unit with different permission")
				err := h.disableJob(job, otherUnitType, timerFile)
				if err != nil {
					return fmt.Errorf("cannot stop or disable existing unit before scheduling with different permission. You might want to retry using sudo.")
				}
				err = h.removeJobFiles(job, otherUnitType, timerFile, systemd.GetServiceFile(job.ProfileName, job.CommandName))
				if err != nil {
					return fmt.Errorf("cannot remove existing unit before scheduling with different permission. You might want to retry using sudo.")
				}
			}
		}
		// tell systemd we've changed some system configuration files
		err := runSystemctlReload(unitType)
		if err != nil {
			return err
		}
	}

	unit := systemd.NewUnit(u)
	err := unit.Generate(systemd.Config{
		CommandLine:          job.Command + " --no-prio " + job.Arguments.String(),
		Environment:          job.Environment,
		WorkingDirectory:     job.WorkingDirectory,
		Title:                job.ProfileName,
		SubTitle:             job.CommandName,
		JobDescription:       job.JobDescription,
		TimerDescription:     job.TimerDescription,
		Schedules:            job.Schedules,
		UnitType:             unitType,
		Priority:             job.GetPriority(),
		UnitFile:             h.config.UnitTemplate,
		TimerFile:            h.config.TimerTemplate,
		AfterNetworkOnline:   job.AfterNetworkOnline,
		DropInFiles:          job.SystemdDropInFiles,
		Nice:                 h.config.Nice,
		IOSchedulingClass:    h.config.IONiceClass,
		IOSchedulingPriority: h.config.IONiceLevel,
		User:                 user,
	})
	if err != nil {
		return err
	}

	if _, callReload := job.GetFlag("reload"); callReload {
		// tell systemd we've changed some system configuration files
		err = runSystemctlReload(unitType)
		if err != nil {
			return err
		}
	}

	extraArgs := []string{"--quiet"}
	if _, noStart := job.GetFlag("no-start"); !noStart {
		// annoyingly, we also have to start it, otherwise it won't be active until the next reboot
		extraArgs = append(extraArgs, "--now")
	}
	// enable (and start) the job
	err = runSystemctlOnUnit(timerFile, systemctlEnable, unitType, false, extraArgs...)
	if err != nil {
		return err
	}

	fmt.Println("")
	// display a status after starting it
	_ = runSystemctlOnUnit(timerFile, systemctlStatus, unitType, false)

	return nil
}

// RemoveJob is disabling the systemd unit and deleting the timer and service files
func (h *HandlerSystemd) RemoveJob(job *Config, permission Permission) error {
	u := user.Current()
	unitType, _ := permissionToSystemd(u, permission)
	serviceFile := systemd.GetServiceFile(job.ProfileName, job.CommandName)
	unitLoaded, err := unitLoaded(serviceFile, unitType)
	if err != nil {
		return err
	}
	if !unitLoaded {
		return ErrScheduledJobNotFound
	}

	timerFile := systemd.GetTimerFile(job.ProfileName, job.CommandName)

	err = h.disableJob(job, unitType, timerFile)
	if err != nil {
		return err
	}

	err = h.removeJobFiles(job, unitType, timerFile, serviceFile)
	if err != nil {
		return err
	}

	if _, callReload := job.GetFlag("reload"); callReload {
		// tell systemd we've changed some system configuration files
		err = runSystemctlReload(unitType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *HandlerSystemd) disableJob(job *Config, unitType systemd.UnitType, timerFile string) error {
	// stop the job with the --now flag then disable the job
	err := runSystemctlOnUnit(timerFile, systemctlDisable, unitType, job.removeOnly, "--now", "--quiet")
	if err != nil {
		return err
	}

	return nil
}

func (h *HandlerSystemd) removeJobFiles(job *Config, unitType systemd.UnitType, timerFile, serviceFile string) error {
	var err error
	unit := systemd.NewUnit(user.Current())
	systemdPath := systemd.GetSystemDir()
	if unitType == systemd.UserUnit {
		systemdPath, err = unit.GetUserDir()
		if err != nil {
			return nil
		}
	}

	dropInDir := systemd.GetServiceFileDropInDir(job.ProfileName, job.CommandName)
	timerDropInDir := systemd.GetTimerFileDropInDir(job.ProfileName, job.CommandName)

	obsoletes := []string{
		path.Join(systemdPath, timerFile),
		path.Join(systemdPath, serviceFile),
		path.Join(systemdPath, timerDropInDir),
		path.Join(systemdPath, dropInDir),
	}

	for _, pathToRemove := range obsoletes {
		if err = os.RemoveAll(pathToRemove); err != nil {
			clog.Errorf("failed removing %q, error: %s. Please remove this path", pathToRemove, err.Error())
		}
	}
	return nil
}

// DisplayJobStatus displays information of a systemd service/timer
func (h *HandlerSystemd) DisplayJobStatus(job *Config) error {
	serviceName := systemd.GetServiceFile(job.ProfileName, job.CommandName)
	timerName := systemd.GetTimerFile(job.ProfileName, job.CommandName)
	permission, _ := h.DetectSchedulePermission(PermissionFromConfig(job.Permission))
	systemdType := systemd.UserUnit
	if permission == PermissionSystem || permission == PermissionUserBackground {
		systemdType = systemd.SystemUnit
	}
	unitLoaded, err := unitLoaded(serviceName, systemdType)
	if err != nil {
		return err
	}
	if !unitLoaded {
		return ErrScheduledJobNotFound
	}
	if systemdType == systemd.UserUnit && user.Current().IsRoot() {
		// journalctl doesn't accept the parameter "-M user@" (yet?)
		clog.Warning("cannot load the journal from a user service as root")
	} else {
		_ = runJournalCtlCommand(timerName, systemdType) // ignore errors on journalctl
	}
	return runSystemctlOnUnit(timerName, systemctlStatus, systemdType, false)
}

func (h *HandlerSystemd) Scheduled(profileName string) ([]Config, error) {
	configs := []Config{}

	cfgs, err := getConfigs(profileName, systemd.SystemUnit)
	if err != nil {
		clog.Errorf("cannot list system units: %s", err)
	}
	if len(cfgs) > 0 {
		configs = append(configs, cfgs...)
	}

	cfgs, err = getConfigs(profileName, systemd.UserUnit)
	if err != nil {
		clog.Errorf("cannot list user units: %s", err)
	}
	if len(cfgs) > 0 {
		configs = append(configs, cfgs...)
	}
	return configs, nil
}

// detectSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission.
// safe specifies whether a guess may lead to a too broad or too narrow file access permission.
func (h *HandlerSystemd) DetectSchedulePermission(p Permission) (Permission, bool) {
	switch p {
	case PermissionSystem, PermissionUserBackground, PermissionUserLoggedOn:
		// well defined
		return p, true

	default:
		// best guess is depending on the user being root or not:
		detected := PermissionUserLoggedOn // sane default
		if os.Geteuid() == 0 {
			detected = PermissionSystem
		}
		// guess based on UID is never safe
		return detected, false
	}
}

// CheckPermission returns true if the user is allowed to access the job.
func (h *HandlerSystemd) CheckPermission(user user.User, p Permission) bool {
	switch p {
	case PermissionUserLoggedOn:
		// user mode is always available
		return true

	default:
		if user.IsRoot() {
			return true
		}
		// last case is system (or undefined) + no sudo
		return false

	}
}

var (
	_ Handler = &HandlerSystemd{}
)

func permissionToSystemd(user user.User, permission Permission) (systemd.UnitType, string) {
	switch permission {
	case PermissionSystem:
		return systemd.SystemUnit, ""

	case PermissionUserBackground:
		return systemd.SystemUnit, user.Username

	case PermissionUserLoggedOn:
		return systemd.UserUnit, ""

	default:
		unitType := systemd.UserUnit
		if user.Uid == 0 {
			unitType = systemd.SystemUnit
		}
		return unitType, ""
	}
}

// getSystemdStatus displays the status of all the timers installed on that profile
func getSystemdStatus(profile string, unitType systemd.UnitType) (string, error) {
	timerName := fmt.Sprintf("resticprofile-*@profile-%s.timer", profile)
	args := []string{"list-timers", "--all", flagNoPager, timerName}
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	buffer := &strings.Builder{}
	cmd, err := systemctlCommand(args...)
	if err != nil {
		return "", err
	}
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	err = cmd.Run()
	return buffer.String(), err
}

func runSystemctlOnUnit(timerName, command string, unitType systemd.UnitType, silent bool, extraArgs ...string) error {
	if command == systemctlStatus {
		fmt.Print("Systemd timer status\n=====================\n")
	}
	args := make([]string, 0, 3)
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	args = append(args, flagNoPager)
	if len(extraArgs) > 0 {
		args = append(args, extraArgs...)
	}
	args = append(args, command, timerName)

	cmd, err := systemctlCommand(args...)
	if err != nil {
		return err
	}
	if !silent {
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
	}
	err = cmd.Run()
	if command == systemctlStatus && cmd.ProcessState.ExitCode() == codeStatusUnitNotFound {
		return ErrScheduledJobNotFound
	}
	if command == systemctlStatus && cmd.ProcessState.ExitCode() == codeStatusNotRunning {
		return ErrScheduledJobNotRunning
	}
	if command == systemctlStop && cmd.ProcessState.ExitCode() == codeStopUnitNotFound {
		return ErrScheduledJobNotFound
	}
	return err
}

func runJournalCtlCommand(timerName string, unitType systemd.UnitType) error {
	fmt.Print("Recent log (>= warning in the last month)\n==========================================\n")
	timerName = strings.TrimSuffix(timerName, ".timer")
	args := []string{"--since", "1 month ago", flagNoPager, "--priority", "warning", "--unit", timerName}
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}

	binary, err := exec.LookPath(journalctlBinary)
	if err != nil {
		return fmt.Errorf("cannot find %q: %w", journalctlBinary, err)
	}
	clog.Debugf("starting command \"%s %s\"", binary, strings.Join(args, " "))
	cmd := exec.Command(binary, args...)
	cmd.Stdout = term.GetOutput()
	cmd.Stderr = term.GetErrorOutput()
	err = cmd.Run()
	fmt.Println("")
	return err
}

func runSystemctlReload(unitType systemd.UnitType) error {
	args := []string{systemctlReload}
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	cmd, err := systemctlCommand(args...)
	if err != nil {
		return err
	}
	cmd.Stdout = term.GetOutput()
	cmd.Stderr = term.GetErrorOutput()
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func listUnits(profile string, unitType systemd.UnitType) ([]SystemdUnit, error) {
	if profile == "" {
		profile = "*"
	}
	pattern := fmt.Sprintf("resticprofile-*@profile-%s.service", profile)
	args := []string{"list-units", "--all", flagNoPager, "--output", "json"}
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	args = append(args, pattern)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd, err := systemctlCommand(args...)
	if err != nil {
		return nil, err
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running command: %w\n%s", err, stderr.String())
	}
	var units []SystemdUnit
	decoder := json.NewDecoder(stdout)
	err = decoder.Decode(&units)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w\n%s", err, stdout.String())
	}
	return units, err
}

func getUserFlags() []string {
	currentUser := user.Current()
	if !currentUser.Sudo {
		return []string{flagUserUnit}
	}
	return []string{flagUserUnit, "-M", currentUser.Username + "@"}
}

func unitLoaded(serviceName string, unitType systemd.UnitType) (bool, error) {
	units, err := listUnits("", unitType)
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(units, func(unit SystemdUnit) bool {
		return unit.Unit == serviceName && unit.Load != unitNotFound
	}), nil
}

func getConfigs(profileName string, unitType systemd.UnitType) ([]Config, error) {
	units, err := listUnits(profileName, unitType)
	if err != nil {
		return nil, err
	}
	configs := make([]Config, 0, len(units))
	for _, unit := range units {
		if unit.Load == unitNotFound {
			continue
		}
		cfg, err := systemd.NewUnit(user.Current()).Read(unit.Unit, unitType)
		if err != nil {
			clog.Errorf("cannot read information from unit %q: %s", unit.Unit, err)
			continue
		}
		if cfg == nil {
			continue
		}
		configs = append(configs, toScheduleConfig(*cfg))
	}
	return configs, nil
}

func toScheduleConfig(systemdConfig systemd.Config) Config {
	var command string
	cmdLine := shell.SplitArguments(systemdConfig.CommandLine)
	if len(cmdLine) > 0 {
		command = cmdLine[0]
	}
	args := NewCommandArguments(cmdLine[1:])

	cfg := Config{
		ConfigFile:       args.ConfigFile(),
		ProfileName:      systemdConfig.Title,
		CommandName:      systemdConfig.SubTitle,
		WorkingDirectory: systemdConfig.WorkingDirectory,
		Command:          command,
		Arguments:        args.Trim([]string{"--no-prio"}),
		JobDescription:   systemdConfig.JobDescription,
		Environment:      systemdConfig.Environment,
		Permission:       systemdConfigPermission(systemdConfig),
		Schedules:        systemdConfig.Schedules,
		Priority:         systemdConfig.Priority,
	}
	return cfg
}

func systemdConfigPermission(systemdConfig systemd.Config) string {
	switch systemdConfig.UnitType {
	case systemd.SystemUnit:
		if systemdConfig.User != "" {
			return constants.SchedulePermissionUser
		}
		return constants.SchedulePermissionSystem
	default:
		return constants.SchedulePermissionUserLoggedOn
	}
}

func systemctlCommand(args ...string) (*exec.Cmd, error) {
	binary, err := exec.LookPath(systemctlBinary)
	if err != nil {
		return nil, fmt.Errorf("cannot find %q: %w", systemctlBinary, err)
	}
	clog.Debugf("starting command \"%s %s\"", binary, strings.Join(args, " "))
	return exec.Command(binary, args...), nil
}

func displaySystemdSchedules(profile, command string, schedules []string) error {
	binary, err := exec.LookPath(analyzeBinary)
	if err != nil {
		return fmt.Errorf("cannot find %q: %w", analyzeBinary, err)
	}

	for index, schedule := range schedules {
		if schedule == "" {
			return errors.New("empty schedule")
		}
		displayHeader(profile, command, index+1, len(schedules))

		cmd := exec.Command(binary, "calendar", schedule)
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	term.Print(platform.LineSeparator)
	return nil
}

// init registers HandlerSystemd
func init() {
	AddHandlerProvider(func(config SchedulerConfig, _ bool) (hr Handler) {
		if config.Type() == constants.SchedulerSystemd ||
			config.Type() == constants.SchedulerOSDefault {
			hr = NewHandlerSystemd(config.Convert(constants.SchedulerSystemd))
		}
		return
	})
}
