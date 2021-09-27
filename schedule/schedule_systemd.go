//+build !darwin,!windows

package schedule

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
)

const (
	journalctlBin  = "journalctl"
	systemctlBin   = "systemctl"
	commandStart   = "start"
	commandStop    = "stop"
	commandEnable  = "enable"
	commandDisable = "disable"
	commandStatus  = "status"
	commandReload  = "daemon-reload"
	flagUserUnit   = "--user"

	// https://www.freedesktop.org/software/systemd/man/systemctl.html#Exit%20status
	codeStatusNotRunning   = 3
	codeStatusUnitNotFound = 4
	codeStopUnitNotFound   = 5 // undocumented
)

// SystemdSchedule is a Scheduler using systemd
type SystemdSchedule struct {
	profileName   string
	unitTemplate  string
	timerTemplate string
}

// Init verifies systemd is available on this system
func (s *SystemdSchedule) Init() error {
	found, err := exec.LookPath(systemctlBin)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like systemd is installed on your system (cannot find %q command in path)", systemctlBin)
	}
	return nil
}

// Close does nothing when using systemd
func (s *SystemdSchedule) Close() {
}

// NewJob instantiates a Job object (of SchedulerJob interface) to schedule jobs
func (s *SystemdSchedule) NewJob(config Config) SchedulerJob {
	return &Job{
		config: config,
		scheduler: SchedulerSystemd{
			UnitTemplate:  s.unitTemplate,
			TimerTemplate: s.timerTemplate,
		},
	}
}

// DisplayStatus display timers in systemd
func (s *SystemdSchedule) DisplayStatus() {
	var (
		status string
		err    error
	)
	if os.Geteuid() == 0 {
		// if the user is root, we search for system timers
		status, err = getSystemdStatus(s.profileName, systemd.SystemUnit)
	} else {
		// otherwise user timers
		status, err = getSystemdStatus(s.profileName, systemd.UserUnit)
	}
	if err != nil || status == "" || strings.HasPrefix(status, "0 timers") {
		// fail silently
		return
	}
	fmt.Printf("\nTimers summary\n===============\n%s\n", status)
}

// Verify interface
var _ Scheduler = &SystemdSchedule{}

// createSystemdJob is creating the systemd unit and activating it
func (j *Job) createSystemdJob(unitType systemd.UnitType) error {
	sch, ok := j.scheduler.(SchedulerSystemd)
	if !ok {
		return errors.New("incompatible scheduler type")
	}
	err := systemd.Generate(systemd.Config{
		CommandLine:      j.config.Command() + " --no-prio " + strings.Join(j.config.Arguments(), " "),
		WorkingDirectory: j.config.WorkingDirectory(),
		Title:            j.config.Title(),
		SubTitle:         j.config.SubTitle(),
		JobDescription:   j.config.JobDescription(),
		TimerDescription: j.config.TimerDescription(),
		Schedules:        j.config.Schedules(),
		UnitType:         unitType,
		Priority:         j.config.Priority(),
		UnitFile:         sch.UnitTemplate,
		TimerFile:        sch.TimerTemplate,
	})
	if err != nil {
		return err
	}

	if unitType == systemd.SystemUnit {
		// tell systemd we've changed some system configuration files
		cmd := exec.Command(systemctlBin, commandReload)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	timerName := systemd.GetTimerFile(j.config.Title(), j.config.SubTitle())

	// enable the job
	err = runSystemctlCommand(timerName, commandEnable, unitType, false)
	if err != nil {
		return err
	}

	if _, noStart := j.config.GetFlag("no-start"); !noStart {
		// annoyingly, we also have to start it, otherwise it won't be active until the next reboot
		err = runSystemctlCommand(timerName, commandStart, unitType, false)
		if err != nil {
			return err
		}
	}
	fmt.Println("")
	// display a status after starting it
	_ = runSystemctlCommand(timerName, commandStatus, unitType, false)

	return nil
}

// removeSystemdJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeSystemdJob(unitType systemd.UnitType) error {
	var err error
	timerFile := systemd.GetTimerFile(j.config.Title(), j.config.SubTitle())

	// stop the job
	err = runSystemctlCommand(timerFile, commandStop, unitType, j.RemoveOnly())
	if err != nil {
		return err
	}

	// disable the job
	err = runSystemctlCommand(timerFile, commandDisable, unitType, j.RemoveOnly())
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

	err = os.Remove(path.Join(systemdPath, timerFile))
	if err != nil {
		return nil
	}

	serviceFile := systemd.GetServiceFile(j.config.Title(), j.config.SubTitle())
	err = os.Remove(path.Join(systemdPath, serviceFile))
	if err != nil {
		return nil
	}

	return nil
}

// displaySystemdStatus displays information of a systemd service/timer
func (j *Job) displaySystemdStatus(command string) error {
	timerName := systemd.GetTimerFile(j.config.Title(), j.config.SubTitle())
	permission := j.getSchedulePermission()
	if permission == constants.SchedulePermissionSystem {
		err := runJournalCtlCommand(timerName, systemd.SystemUnit)
		if err != nil {
			clog.Warningf("cannot read system logs: %v", err)
		}
		return runSystemctlCommand(timerName, commandStatus, systemd.SystemUnit, false)
	}
	err := runJournalCtlCommand(timerName, systemd.UserUnit)
	if err != nil {
		clog.Warningf("cannot read user logs: %v", err)
	}
	return runSystemctlCommand(timerName, commandStatus, systemd.UserUnit, false)
}

// getSystemdStatus displays the status of all the timers installed on that profile
func getSystemdStatus(profile string, unitType systemd.UnitType) (string, error) {
	timerName := fmt.Sprintf("resticprofile-*@profile-%s.timer", profile)
	args := []string{"list-timers", "--all", "--no-pager", timerName}
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	buffer := &strings.Builder{}
	cmd := exec.Command(systemctlBin, args...)
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	err := cmd.Run()
	return buffer.String(), err
}

func runSystemctlCommand(timerName, command string, unitType systemd.UnitType, silent bool) error {
	if command == commandStatus {
		fmt.Print("Systemd timer status\n=====================\n")
	}
	args := make([]string, 0, 3)
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	args = append(args, command, timerName)

	cmd := exec.Command(systemctlBin, args...)
	if !silent {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if command == commandStatus && cmd.ProcessState.ExitCode() == codeStatusUnitNotFound {
		return ErrorServiceNotFound
	}
	if command == commandStatus && cmd.ProcessState.ExitCode() == codeStatusNotRunning {
		return ErrorServiceNotRunning
	}
	if command == commandStop && cmd.ProcessState.ExitCode() == codeStopUnitNotFound {
		return ErrorServiceNotFound
	}
	return err
}

func runJournalCtlCommand(timerName string, unitType systemd.UnitType) error {
	fmt.Print("Recent log (>= warning in the last month)\n==========================================\n")
	timerName = strings.TrimSuffix(timerName, ".timer")
	args := []string{"--since", "1 month ago", "--no-pager", "--priority", "warning", "--unit", timerName}
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	cmd := exec.Command(journalctlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Println("")
	return err
}
