//go:build !darwin && !windows

package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
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
	journalctlBinary = "/usr/bin/journalctl"
	systemctlBinary  = "/usr/bin/systemctl"
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
	unitType, user := permissionToSystemd(permission)

	if unitType == systemd.UserUnit && job.AfterNetworkOnline {
		return fmt.Errorf("after-network-online is not available for \"user_logged_on\" permission schedules")
	}

	err := systemd.Generate(systemd.Config{
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

	// tell systemd we've changed some system configuration files
	err = runSystemctlReload(unitType)
	if err != nil {
		return err
	}

	timerName := systemd.GetTimerFile(job.ProfileName, job.CommandName)

	// enable the job
	err = runSystemctlCommand(timerName, systemctlEnable, unitType, false)
	if err != nil {
		return err
	}

	if _, noStart := job.GetFlag("no-start"); !noStart {
		// annoyingly, we also have to start it, otherwise it won't be active until the next reboot
		err = runSystemctlCommand(timerName, systemctlStart, unitType, false)
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	// display a status after starting it
	_ = runSystemctlCommand(timerName, systemctlStatus, unitType, false)

	return nil
}

// RemoveJob is disabling the systemd unit and deleting the timer and service files
func (h *HandlerSystemd) RemoveJob(job *Config, permission Permission) error {
	var err error
	unitType, _ := permissionToSystemd(permission)
	serviceFile := systemd.GetServiceFile(job.ProfileName, job.CommandName)
	unitLoaded, err := unitLoaded(serviceFile, unitType)
	if err != nil {
		return err
	}
	if !unitLoaded {
		return ErrScheduledJobNotFound
	}

	timerFile := systemd.GetTimerFile(job.ProfileName, job.CommandName)

	// stop the job
	err = runSystemctlCommand(timerFile, systemctlStop, unitType, job.removeOnly)
	if err != nil {
		return err
	}

	// disable the job
	err = runSystemctlCommand(timerFile, systemctlDisable, unitType, job.removeOnly)
	if err != nil {
		return err
	}

	systemdPath := systemd.GetSystemDir()
	if unitType == systemd.UserUnit {
		systemdPath, err = systemd.GetUserDir()
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

	// tell systemd we've changed some system configuration files
	err = runSystemctlReload(unitType)
	if err != nil {
		return err
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
	_ = runJournalCtlCommand(timerName, systemdType) // ignore errors on journalctl
	return runSystemctlCommand(timerName, systemctlStatus, systemdType, false)
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
func (h *HandlerSystemd) CheckPermission(p Permission) bool {
	switch p {
	case PermissionUserLoggedOn:
		// user mode is always available
		return true

	default:
		if os.Geteuid() == 0 {
			// user has sudoed
			return true
		}
		// last case is system (or undefined) + no sudo
		return false

	}
}

var (
	_ Handler = &HandlerSystemd{}
)

func permissionToSystemd(permission Permission) (systemd.UnitType, string) {
	switch permission {
	case PermissionSystem:
		return systemd.SystemUnit, ""

	case PermissionUserBackground:
		return systemd.SystemUnit, user.Current().Username

	case PermissionUserLoggedOn:
		return systemd.UserUnit, ""

	default:
		unitType := systemd.UserUnit
		if os.Geteuid() == 0 {
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
	cmd := systemctlCommand(args...)
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	err := cmd.Run()
	return buffer.String(), err
}

func runSystemctlCommand(timerName, command string, unitType systemd.UnitType, silent bool) error {
	if command == systemctlStatus {
		fmt.Print("Systemd timer status\n=====================\n")
	}
	args := make([]string, 0, 3)
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	args = append(args, flagNoPager)
	args = append(args, command, timerName)

	cmd := systemctlCommand(args...)
	if !silent {
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
	}
	err := cmd.Run()
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
	clog.Debugf("starting command \"%s %s\"", journalctlBinary, strings.Join(args, " "))
	cmd := exec.Command(journalctlBinary, args...)
	cmd.Stdout = term.GetOutput()
	cmd.Stderr = term.GetErrorOutput()
	err := cmd.Run()
	fmt.Println("")
	return err
}

func runSystemctlReload(unitType systemd.UnitType) error {
	args := []string{systemctlReload}
	if unitType == systemd.UserUnit {
		args = append(args, getUserFlags()...)
	}
	cmd := systemctlCommand(args...)
	cmd.Stdout = term.GetOutput()
	cmd.Stderr = term.GetErrorOutput()
	err := cmd.Run()
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
	cmd := systemctlCommand(args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
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
	if !currentUser.SudoRoot {
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
		cfg, err := systemd.Read(unit.Unit, unitType)
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

func systemctlCommand(args ...string) *exec.Cmd {
	clog.Debugf("starting command \"%s %s\"", systemctlBinary, strings.Join(args, " "))
	return exec.Command(systemctlBinary, args...)
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
