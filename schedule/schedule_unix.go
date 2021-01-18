//+build !darwin,!windows

package schedule

//
// Schedule: common code for systemd and crond only
//

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
)

// Init verifies systemd or crond is available on this system
func Init() error {
	scheduler := constants.SchedulerSystemd
	bin := systemctlBin
	if Scheduler == constants.SchedulerCrond {
		scheduler = constants.SchedulerCrond
		bin = crontabBin
	}
	found, err := exec.LookPath(bin)
	if err != nil || found == "" {
		return fmt.Errorf("it doesn't look like %s is installed on your system (cannot find %q command in path)", scheduler, bin)
	}
	return nil
}

// Close does nothing in systemd or crond
func Close() {
}

// createJob is creating the crontab OR systemd unit and activating it
func (j *Job) createJob(schedules []*calendar.Event) error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to create a system job: please restart resticprofile as root (with sudo)")
	}
	if Scheduler == constants.SchedulerCrond {
		return j.createCrondJob(schedules)
	}
	if os.Geteuid() == 0 {
		// user has sudoed already
		return j.createSystemdJob(systemd.SystemUnit)
	}
	return j.createSystemdJob(systemd.UserUnit)
}

// removeJob is disabling the systemd unit and deleting the timer and service files
func (j *Job) removeJob() error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to remove a system job: please restart resticprofile as root (with sudo)")
	}
	if Scheduler == constants.SchedulerCrond {
		return j.removeCrondJob()
	}
	if os.Geteuid() == 0 {
		// user has sudoed
		return j.removeSystemdJob(systemd.SystemUnit)
	}
	return j.removeSystemdJob(systemd.UserUnit)
}

// displayStatus of a schedule
func (j *Job) displayStatus(command string) error {
	if Scheduler == constants.SchedulerCrond {
		return j.displayCrondStatus(command)
	}
	return j.displaySystemdStatus(command)
}
