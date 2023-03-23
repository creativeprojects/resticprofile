package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
)

func scheduleJobs(handler schedule.Handler, profileName string, configs []*config.ScheduleConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	scheduler := schedule.NewScheduler(handler, profileName)
	err = scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, scheduleConfig := range configs {
		args := []string{
			"--no-ansi",
			"--config",
			scheduleConfig.ConfigFile,
			"--name",
			scheduleConfig.Title,
		}

		if scheduleConfig.Log != "" {
			args = append(args, "--log", scheduleConfig.Log)
		}

		if scheduleConfig.GetLockMode() == config.ScheduleLockModeDefault {
			if scheduleConfig.GetLockWait() > 0 {
				args = append(args, "--lock-wait", scheduleConfig.GetLockWait().String())
			}
		} else if scheduleConfig.GetLockMode() == config.ScheduleLockModeIgnore {
			args = append(args, "--no-lock")
		}

		args = append(args, getResticCommand(scheduleConfig.SubTitle))

		scheduleConfig.SetCommand(wd, binary, args)
		scheduleConfig.JobDescription =
			fmt.Sprintf("resticprofile %s for profile %s in %s", scheduleConfig.SubTitle, scheduleConfig.Title, scheduleConfig.ConfigFile)
		scheduleConfig.TimerDescription =
			fmt.Sprintf("%s timer for profile %s in %s", scheduleConfig.SubTitle, scheduleConfig.Title, scheduleConfig.ConfigFile)

		job := scheduler.NewJob(scheduleConfig)
		err = job.Create()
		if err != nil {
			return fmt.Errorf("error creating job %s/%s: %w",
				scheduleConfig.Title,
				scheduleConfig.SubTitle,
				err)
		}
		clog.Infof("scheduled job %s/%s created", scheduleConfig.Title, scheduleConfig.SubTitle)
	}
	return nil
}

func removeJobs(handler schedule.Handler, profileName string, configs []*config.ScheduleConfig) error {
	scheduler := schedule.NewScheduler(handler, profileName)
	err := scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, scheduleConfig := range configs {
		job := scheduler.NewJob(scheduleConfig)

		// Skip over non-accessible, RemoveOnly jobs since they may not exist and must not causes errors
		if job.RemoveOnly() && !job.Accessible() {
			continue
		}

		// Try to remove the job
		err := job.Remove()
		if err != nil {
			if errors.Is(err, schedule.ErrorServiceNotFound) {
				// Display a warning and keep going. Skip message for RemoveOnly jobs since they may not exist
				if !job.RemoveOnly() {
					clog.Warningf("service %s/%s not found", scheduleConfig.Title, scheduleConfig.SubTitle)
				}
				continue
			}
			return fmt.Errorf("error removing job %s/%s: %w",
				scheduleConfig.Title,
				scheduleConfig.SubTitle,
				err)
		}

		clog.Infof("scheduled job %s/%s removed", scheduleConfig.Title, scheduleConfig.SubTitle)
	}
	return nil
}

func statusJobs(handler schedule.Handler, profileName string, configs []*config.ScheduleConfig) error {
	scheduler := schedule.NewScheduler(handler, profileName)
	err := scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, scheduleConfig := range configs {
		job := scheduler.NewJob(scheduleConfig)
		err := job.Status()
		if err != nil {
			if errors.Is(err, schedule.ErrorServiceNotFound) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s not found", scheduleConfig.Title, scheduleConfig.SubTitle)
				continue
			}
			if errors.Is(err, schedule.ErrorServiceNotRunning) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s is not running", scheduleConfig.Title, scheduleConfig.SubTitle)
				continue
			}
			return fmt.Errorf("error querying status of job %s/%s: %w",
				scheduleConfig.Title,
				scheduleConfig.SubTitle,
				err)
		}
	}
	scheduler.DisplayStatus()
	return nil
}

func getResticCommand(profileCommand string) string {
	if profileCommand == constants.SectionConfigurationRetention {
		return constants.CommandForget
	}
	return profileCommand
}
