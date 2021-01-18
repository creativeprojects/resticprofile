//+build !darwin,!windows

package schedule

import (
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

// createSystemdJob is creating the systemd unit and activating it
func (j *Job) createSystemdJob(unitType systemd.UnitType) error {
	err := systemd.Generate(
		j.config.Command()+" --no-prio "+strings.Join(j.config.Arguments(), " "),
		j.config.WorkingDirectory(),
		j.config.Title(),
		j.config.SubTitle(),
		j.config.JobDescription(),
		j.config.TimerDescription(),
		j.config.Schedules(),
		unitType,
		j.config.Priority())
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
	err = runSystemctlCommand(timerName, commandEnable, unitType)
	if err != nil {
		return err
	}

	// annoyingly, we also have to start it, otherwise it won't be active until the next reboot
	err = runSystemctlCommand(timerName, commandStart, unitType)
	if err != nil {
		return err
	}
	// display a status after starting it
	_ = runSystemctlCommand(timerName, commandStatus, unitType)

	return nil
}

// removeSystemdJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeSystemdJob(unitType systemd.UnitType) error {
	var err error
	timerFile := systemd.GetTimerFile(j.config.Title(), j.config.SubTitle())

	// stop the job
	err = runSystemctlCommand(timerFile, commandStop, unitType)
	if err != nil {
		return err
	}

	// disable the job
	err = runSystemctlCommand(timerFile, commandDisable, unitType)
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
		return runSystemctlCommand(timerName, commandStatus, systemd.SystemUnit)
	}
	err := runJournalCtlCommand(timerName, systemd.UserUnit)
	if err != nil {
		clog.Warningf("cannot read user logs: %v", err)
	}
	return runSystemctlCommand(timerName, commandStatus, systemd.UserUnit)
}

func runSystemctlCommand(timerName, command string, unitType systemd.UnitType) error {
	args := make([]string, 0, 3)
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	args = append(args, command, timerName)

	cmd := exec.Command(systemctlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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
	timerName = strings.TrimSuffix(timerName, ".timer")
	args := []string{"--since", "1 month ago", "--no-pager", "--priority", "err", "--unit", timerName}
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
