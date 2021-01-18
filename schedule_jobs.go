package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
)

func scheduleJobs(configs []*config.ScheduleConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	err = schedule.Init()
	if err != nil {
		return err
	}
	defer schedule.Close()

	for _, scheduleConfig := range configs {
		args := []string{
			"--no-ansi",
			"--config",
			scheduleConfig.Configfile(),
			"--name",
			scheduleConfig.Title(),
		}
		if runtime.GOOS != "darwin" && scheduleConfig.Logfile() != "" {
			args = append(args, "--log", scheduleConfig.Logfile())
		}
		args = append(args, getResticCommand(scheduleConfig.SubTitle()))

		scheduleConfig.SetCommand(wd, binary, args)
		scheduleConfig.SetJobDescription(
			fmt.Sprintf("resticprofile %s for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), scheduleConfig.Configfile()))
		scheduleConfig.SetTimerDescription(
			fmt.Sprintf("%s timer for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), scheduleConfig.Configfile()))

		job := schedule.NewJob(scheduleConfig)
		err = job.Create()
		if err != nil {
			return fmt.Errorf("error creating job %s/%s: %w",
				scheduleConfig.Title(),
				scheduleConfig.SubTitle(),
				err)
		}
		clog.Infof("scheduled job %s/%s created", scheduleConfig.Title(), scheduleConfig.SubTitle())
	}
	return nil
}

func removeJobs(configs []*config.ScheduleConfig) error {
	err := schedule.Init()
	if err != nil {
		return err
	}
	defer schedule.Close()

	for _, scheduleConfig := range configs {
		job := schedule.NewJob(scheduleConfig)
		err := job.Remove()
		if err != nil {
			if errors.Is(err, schedule.ErrorServiceNotFound) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s not found", scheduleConfig.Title(), scheduleConfig.SubTitle())
				continue
			}
			return fmt.Errorf("error removing job %s/%s: %w",
				scheduleConfig.Title(),
				scheduleConfig.SubTitle(),
				err)
		}
		clog.Infof("scheduled job %s/%s removed", scheduleConfig.Title(), scheduleConfig.SubTitle())
	}
	return nil
}

func statusJobs(configs []*config.ScheduleConfig) error {
	err := schedule.Init()
	if err != nil {
		return err
	}
	defer schedule.Close()

	for _, scheduleConfig := range configs {
		job := schedule.NewJob(scheduleConfig)
		err := job.Status()
		if err != nil {
			if errors.Is(err, schedule.ErrorServiceNotFound) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s not found", scheduleConfig.Title(), scheduleConfig.SubTitle())
				continue
			}
			if errors.Is(err, schedule.ErrorServiceNotRunning) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s is not running", scheduleConfig.Title(), scheduleConfig.SubTitle())
				continue
			}
			return fmt.Errorf("error querying status of job %s/%s: %w",
				scheduleConfig.Title(),
				scheduleConfig.SubTitle(),
				err)
		}
	}
	return nil
}

func getResticCommand(profileCommand string) string {
	if profileCommand == constants.SectionConfigurationRetention {
		return constants.CommandForget
	}
	return profileCommand
}
