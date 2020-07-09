//+build windows

package schedule

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/win"
)

// createJob is creating the task scheduler job.
func (j *Job) createJob() error {
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	args := fmt.Sprintf("--no-ansi --config %s --name %s backup", j.configFile, j.profile.Name)
	description := fmt.Sprintf("restic backup using profile '%s' from '%s'", j.profile.Name, j.configFile)

	// default permission will be system
	permission := win.SystemAccount
	if j.profile.Backup.SchedulePermission == constants.SchedulePermissionUser {
		permission = win.UserAccount
	}
	taskScheduler := win.NewTaskScheduler(j.profile)
	err = taskScheduler.Create(binary, args, wd, description, j.schedules, permission)
	if err != nil {
		return err
	}
	return nil
}

// removeJob is deleting the task scheduler job
func (j *Job) removeJob() error {
	taskScheduler := win.NewTaskScheduler(j.profile)
	err := taskScheduler.Delete()
	if err != nil {
		return err
	}
	return nil
}

// checkSystem does nothing on windows as the task scheduler is always available
func checkSystem() error {
	return nil
}

// displayStatus display some information about the task scheduler job
func (j *Job) displayStatus() error {
	taskScheduler := win.NewTaskScheduler(j.profile)
	err := taskScheduler.Status()
	if err != nil {
		return err
	}
	return nil
}
