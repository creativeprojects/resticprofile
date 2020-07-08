//+build !darwin,!windows

package schedule

import (
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
	"github.com/creativeprojects/resticprofile/term"
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
)

// checkSystem verifies systemd is available on this system
func checkSystem() error {
	found, err := exec.LookPath(systemdBin)
	if err != nil || found == "" {
		return errors.New("it doesn't look like systemd is installed on your system")
	}
	return nil
}

func (j *Job) checkPermission() bool {
	if j.profile.SchedulePermission == constants.SchedulePermissionUser {
		// user mode is always available
		return true
	}
	if os.Geteuid() == 0 {
		// user has sudo'ed
		return true
	}
	// last case is system (or undefined) + no sudo
	return false
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeJob() error {
	if os.Geteuid() == 0 {
		// user has sudoed
		return j.removeSystemdJob(systemd.SystemUnit)
	}
	return j.removeSystemdJob(systemd.UserUnit)
}

// removeSystemdJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeSystemdJob(unitType systemd.UnitType) error {
	var err error

	// stop the job
	err = runSystemdCommand(j.profile.Name, commandStop, unitType)
	if err != nil {
		return err
	}

	// disable the job
	err = runSystemdCommand(j.profile.Name, commandDisable, unitType)
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
	timerFile := systemd.GetTimerFile(j.profile.Name)
	err = os.Remove(path.Join(systemdPath, timerFile))
	if err != nil {
		return nil
	}

	serviceFile := systemd.GetServiceFile(j.profile.Name)
	err = os.Remove(path.Join(systemdPath, serviceFile))
	if err != nil {
		return nil
	}

	return nil
}

// createJob is creating the systemd unit and activating it
func (j *Job) createJob() error {
	if os.Geteuid() == 0 {
		// user has sudoed already
		return j.createSystemdJob(systemd.SystemUnit)
	}
	message := "\nPlease note resticprofile was started as a standard user (typically without sudo):" +
		"\nDo you want to install the scheduled backup as a user job as opposed to a system job?"
	answer := term.AskYesNo(os.Stdin, message, false)
	if !answer {
		return errors.New("operation cancelled by user")
	}
	return j.createSystemdJob(systemd.UserUnit)
}

func (j *Job) createSystemdJob(unitType systemd.UnitType) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	err = systemd.Generate(wd, binary, j.configFile, j.profile.Name, j.profile.Schedule, unitType)
	if err != nil {
		return err
	}

	// enable the job
	err = runSystemdCommand(j.profile.Name, commandEnable, unitType)
	if err != nil {
		return err
	}

	// start the job
	err = runSystemdCommand(j.profile.Name, commandStart, unitType)
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) displayStatus() error {
	if os.Geteuid() == 0 {
		// user has sudoed
		return runSystemdCommand(j.profile.Name, commandStatus, systemd.SystemUnit)
	}
	return runSystemdCommand(j.profile.Name, commandStatus, systemd.UserUnit)
}

func runSystemdCommand(profileName, command string, unitType systemd.UnitType) error {
	args := make([]string, 0, 3)
	if unitType == systemd.UserUnit {
		args = append(args, flagUserUnit)
	}
	args = append(args, command, systemd.GetTimerFile(profileName))

	cmd := exec.Command(systemctlBin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
