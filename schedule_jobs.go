package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/schedule"
)

func scheduleJobs(configFile string, configs []*config.ScheduleConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	for _, scheduleConfig := range configs {
		scheduleConfig.SetCommand(wd, binary, []string{
			"--no-ansi",
			"--config",
			configFile,
			"--name",
			scheduleConfig.Title(),
			getResticCommand(scheduleConfig.SubTitle()),
		})
		scheduleConfig.SetJobDescription(
			fmt.Sprintf("resticprofile %s for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), configFile))
		scheduleConfig.SetTimerDescription(
			fmt.Sprintf("%s timer for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), configFile))

		job := schedule.NewJob(scheduleConfig)
		err = job.Create()
		if err != nil {
			return err
		}
		clog.Infof("scheduled job %s/%s created", scheduleConfig.Title(), scheduleConfig.SubTitle())
	}
	return nil
}

func removeJobs(configs []*config.ScheduleConfig) error {
	for _, scheduleConfig := range configs {
		job := schedule.NewJob(scheduleConfig)
		err := job.Remove()
		if err != nil {
			return err
		}
		clog.Infof("scheduled job %s/%s removed", scheduleConfig.Title(), scheduleConfig.SubTitle())
	}
	return nil
}

func statusJobs(configs []*config.ScheduleConfig) error {
	for _, scheduleConfig := range configs {
		job := schedule.NewJob(scheduleConfig)
		err := job.Status()
		if err != nil {
			return err
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
