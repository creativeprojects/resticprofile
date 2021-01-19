//+build !darwin,!windows

package schedule

//
// Schedule: common code for systemd and crond only
//

import (
	"errors"
	"os"

	"github.com/creativeprojects/resticprofile/calendar"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/systemd"
)

// NewScheduler creates a Scheduler interface, which is either a CrondSchedule or a SystemdSchedule object
func NewScheduler(scheduler, profileName string) Scheduler {
	if scheduler == constants.SchedulerCrond {
		return &CrondSchedule{
			profileName: profileName,
		}
	}
	return &SystemdSchedule{
		profileName: profileName,
	}
}

// createJob is creating the crontab OR systemd unit and activating it
func (j *Job) createJob(schedules []*calendar.Event) error {
	permission := j.getSchedulePermission()
	ok := j.checkPermission(permission)
	if !ok {
		return errors.New("user is not allowed to create a system job: please restart resticprofile as root (with sudo)")
	}
	if j.scheduler == constants.SchedulerCrond {
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
	if j.scheduler == constants.SchedulerCrond {
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
	if j.scheduler == constants.SchedulerCrond {
		return j.displayCrondStatus(command)
	}
	return j.displaySystemdStatus(command)
}
