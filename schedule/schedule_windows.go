//+build windows

package schedule

import (
	"fmt"
	"os"

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

	taskScheduler := win.NewTaskScheduler(j.profile)
	err = taskScheduler.Create(binary, args, wd, description)
	if err != nil {
		return retryElevated(err)
	}
	return nil
}

func (j *Job) removeJob() error {
	taskScheduler := win.NewTaskScheduler(j.profile)
	err := taskScheduler.Delete()
	if err != nil {
		return retryElevated(err)
	}
	return nil
}

// checkSystem does nothing on windows as the task scheduler is always available
func checkSystem() error {
	return nil
}

func (j *Job) displayStatus() error {
	taskScheduler := win.NewTaskScheduler(j.profile)
	err := taskScheduler.Status()
	if err != nil {
		return retryElevated(err)
	}
	return nil
}

func retryElevated(err error) error {
	if err == nil {
		return nil
	}
	// maybe can find a better way than searching for the word "denied"?
	// if strings.Contains(err.Error(), "denied") {
	// 	clog.Info("restarting resticprofile in elevated mode...")
	// 	err := win.RunElevated()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }
	return err
}
