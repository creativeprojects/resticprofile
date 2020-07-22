//+build !darwin,!windows

package schedule

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
)

const (
	systemdBin     = "systemd"
	systemctlBin   = "systemctl"
	commandStart   = "start"
	commandStop    = "stop"
	commandEnable  = "enable"
	commandDisable = "disable"
	commandStatus  = "status"
	flagUserUnit   = "--user"

	// https://www.freedesktop.org/software/systemd/man/systemctl.html#Exit%20status
	codeUnitNotFound = 4
)

// Init verifies systemd is available on this system
func Init() error {
	found, err := exec.LookPath(systemdBin)
	if err != nil || found == "" {
		return errors.New("it doesn't look like systemd is installed on your system")
	}
	return nil
}

// Close does nothing in systemd
func Close() {
}

// getSchedulePermission returns the permission defined from the configuration,
// or the best guess considering the current user permission
func (j *Job) getSchedulePermission() string {
	const message = "you have not specified the permission for your schedule (system or user): assuming"
	if j.config.Permission() == constants.SchedulePermissionSystem ||
		j.config.Permission() == constants.SchedulePermissionUser {
		// well defined
		return j.config.Permission()
	}
	// best guess is depending on the user being root or not:
	if os.Geteuid() == 0 {
		clog.Warning(message, "system")
		return constants.SchedulePermissionSystem
	}
	clog.Warning(message, "user")
	return constants.SchedulePermissionUser
}

func (j *Job) checkPermission(permission string) bool {
	if permission == constants.SchedulePermissionUser {
		// user mode is always available
		return true
	}
	if os.Geteuid() == 0 {
		// user has sudoed
		return true
	}
	// last case is system (or undefined) + no sudo
	return false
}

// createJob is creating the systemd unit and activating it
func (j *Job) createJob(schedules []*calendar.Event) error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to create a system job: please restart resticprofile as root (with sudo)")
	}
	if os.Geteuid() == 0 {
		// user has sudoed already
		return j.createSystemdJob(systemd.SystemUnit)
	}
	return j.createSystemdJob(systemd.UserUnit)
}

// createSystemdJob is creating the systemd unit and activating it
func (j *Job) createSystemdJob(unitType systemd.UnitType) error {

	err := systemd.Generate(
		j.config.Command()+" "+strings.Join(j.config.Arguments(), " "),
		j.config.WorkingDirectory(),
		j.config.Title(),
		j.config.SubTitle(),
		j.config.JobDescription(),
		j.config.TimerDescription(),
		j.config.Schedules(),
		unitType)
	if err != nil {
		return err
	}

	timerName := systemd.GetTimerFile(j.config.Title(), j.config.SubTitle())

	// enable the job
	err = runSystemctlCommand(timerName, commandEnable, unitType)
	if err != nil {
		return err
	}

	// start the job
	err = runSystemctlCommand(timerName, commandStart, unitType)
	if err != nil {
		return err
	}

	return nil
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeJob() error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to remove a system job: please restart resticprofile as root (with sudo)")
	}
	if os.Geteuid() == 0 {
		// user has sudoed
		return j.removeSystemdJob(systemd.SystemUnit)
	}
	return j.removeSystemdJob(systemd.UserUnit)
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

// displayStatus of a systemd service/timer
func (j *Job) displayStatus(command string) error {
	permission := j.getSchedulePermission()
	if permission == constants.SchedulePermissionSystem {
		return runSystemctlCommand(systemd.GetTimerFile(j.config.Title(), j.config.SubTitle()), commandStatus, systemd.SystemUnit)
	}
	return runSystemctlCommand(systemd.GetTimerFile(j.config.Title(), j.config.SubTitle()), commandStatus, systemd.UserUnit)
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
	if cmd.ProcessState.ExitCode() == codeUnitNotFound {
		return ErrorServiceNotFound
	}
	return err
}
