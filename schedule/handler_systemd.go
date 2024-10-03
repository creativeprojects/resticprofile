//go:build !darwin && !windows

package schedule

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/term"
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

	// https://www.freedesktop.org/software/systemd/man/systemctl.html#Exit%20status
	codeStatusNotRunning   = 3
	codeStatusUnitNotFound = 4
	codeStopUnitNotFound   = 5 // undocumented
)

var (
	journalctlBinary = "journalctl"
	systemctlBinary  = "systemctl"
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

// DisplayParsedSchedules does nothing with systemd
func (h *HandlerSystemd) DisplayParsedSchedules(command string, events []*calendar.Event) {}

// DisplaySchedules displays the schedules through the systemd-analyze command
func (h *HandlerSystemd) DisplaySchedules(command string, schedules []string) error {
	return displaySystemdSchedules(command, schedules)
}

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
	if err != nil || status == "" || strings.HasPrefix(status, "0 timers") {
		// fail silently
		return nil
	}
	fmt.Fprintf(term.GetOutput(), "\nTimers summary\n===============\n%s\n", status)
	return nil
}

// CreateJob is creating the systemd unit and activating it
func (h *HandlerSystemd) CreateJob(job *Config, schedules []*calendar.Event, permission string) error {
	unitType := systemd.UserUnit
	if os.Geteuid() == 0 {
		// user has sudoed already
		unitType = systemd.SystemUnit
	}

	if unitType == systemd.UserUnit && job.AfterNetworkOnline {
		return fmt.Errorf("after-network-online only available for \"system\" permission schedules")
	}

	err := systemd.Generate(systemd.Config{
		CommandLine:        job.Command + " --no-prio " + strings.Join(job.Arguments, " "),
		Environment:        job.Environment,
		WorkingDirectory:   job.WorkingDirectory,
		Title:              job.ProfileName,
		SubTitle:           job.CommandName,
		JobDescription:     job.JobDescription,
		TimerDescription:   job.TimerDescription,
		Schedules:          job.Schedules,
		UnitType:           unitType,
		Priority:           job.GetPriority(),
		UnitFile:           h.config.UnitTemplate,
		TimerFile:          h.config.TimerTemplate,
		AfterNetworkOnline: job.AfterNetworkOnline,
		DropInFiles:        job.SystemdDropInFiles,
		Nice:               job.Nice,
	})
	if err != nil {
		return err
	}

	if unitType == systemd.SystemUnit {
		// tell systemd we've changed some system configuration files
		cmd := exec.Command(systemctlBinary, systemctlReload)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
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
func (h *HandlerSystemd) RemoveJob(job *Config, permission string) error {
	unitType := systemd.UserUnit
	if os.Geteuid() == 0 {
		// user has sudoed already
		unitType = systemd.SystemUnit
	}
	var err error
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

	serviceFile := systemd.GetServiceFile(job.ProfileName, job.CommandName)
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
	timerName := systemd.GetTimerFile(job.ProfileName, job.CommandName)
	permission := getSchedulePermission(job.Permission)
	if permission == constants.SchedulePermissionSystem {
		err := runJournalCtlCommand(timerName, systemd.SystemUnit)
		if err != nil {
			clog.Warningf("cannot read system logs: %v", err)
		}
		return runSystemctlCommand(timerName, systemctlStatus, systemd.SystemUnit, false)
	}
	err := runJournalCtlCommand(timerName, systemd.UserUnit)
	if err != nil {
		clog.Warningf("cannot read user logs: %v", err)
	}
	return runSystemctlCommand(timerName, systemctlStatus, systemd.UserUnit, false)
}

var (
	_ Handler = &HandlerSystemd{}
)

// getSystemdStatus displays the status of all the timers installed on that profile
func getSystemdStatus(profile string, unitType systemd.UnitType) (string, error) {
	timerName := fmt.Sprintf("resticprofile-*@profile-%s.timer", profile)
	args := []string{"list-timers", "--all", flagNoPager, timerName}
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	clog.Debugf("starting command \"%s %s\"", systemctlBinary, strings.Join(args, " "))
	buffer := &strings.Builder{}
	cmd := exec.Command(systemctlBinary, args...)
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
		args = append(args, flagUserUnit)
	}
	args = append(args, flagNoPager)
	args = append(args, command, timerName)

	clog.Debugf("starting command \"%s %s\"", systemctlBinary, strings.Join(args, " "))
	cmd := exec.Command(systemctlBinary, args...)
	if !silent {
		cmd.Stdout = term.GetOutput()
		cmd.Stderr = term.GetErrorOutput()
	}
	err := cmd.Run()
	if command == systemctlStatus && cmd.ProcessState.ExitCode() == codeStatusUnitNotFound {
		return ErrServiceNotFound
	}
	if command == systemctlStatus && cmd.ProcessState.ExitCode() == codeStatusNotRunning {
		return ErrServiceNotRunning
	}
	if command == systemctlStop && cmd.ProcessState.ExitCode() == codeStopUnitNotFound {
		return ErrServiceNotFound
	}
	return err
}

func runJournalCtlCommand(timerName string, unitType systemd.UnitType) error {
	fmt.Print("Recent log (>= warning in the last month)\n==========================================\n")
	timerName = strings.TrimSuffix(timerName, ".timer")
	args := []string{"--since", "1 month ago", flagNoPager, "--priority", "warning", "--unit", timerName}
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	clog.Debugf("starting command \"%s %s\"", journalctlBinary, strings.Join(args, " "))
	cmd := exec.Command(journalctlBinary, args...)
	cmd.Stdout = term.GetOutput()
	cmd.Stderr = term.GetErrorOutput()
	err := cmd.Run()
	fmt.Println("")
	return err
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
