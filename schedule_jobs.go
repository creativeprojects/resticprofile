package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/schedule"
)

func scheduleJobs(schedulerType schedule.SchedulerConfig, profileName string, configs []*config.ScheduleConfig) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}

	scheduler := schedule.NewScheduler(schedulerType, profileName)
	err = scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, scheduleConfig := range configs {
		args := []string{
			"--no-ansi",
			"--config",
			scheduleConfig.Configfile(),
			"--name",
			scheduleConfig.Title(),
		}

		if !platform.IsDarwin() && scheduleConfig.Logfile() != "" {
			args = append(args, "--log", scheduleConfig.Logfile())
		}

		if !platform.IsWindows() && scheduleConfig.Syslog() != "" {
			args = append(args, "--syslog", scheduleConfig.Syslog())
		}

		if scheduleConfig.LockMode() == config.ScheduleLockModeDefault {
			if scheduleConfig.LockWait() > 0 {
				args = append(args, "--lock-wait", scheduleConfig.LockWait().String())
			}
		} else if scheduleConfig.LockMode() == config.ScheduleLockModeIgnore {
			args = append(args, "--no-lock")
		}

		args = append(args, getResticCommand(scheduleConfig.SubTitle()))

		scheduleConfig.SetCommand(wd, binary, args)
		scheduleConfig.SetJobDescription(
			fmt.Sprintf("resticprofile %s for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), scheduleConfig.Configfile()))
		scheduleConfig.SetTimerDescription(
			fmt.Sprintf("%s timer for profile %s in %s", scheduleConfig.SubTitle(), scheduleConfig.Title(), scheduleConfig.Configfile()))

		job := scheduler.NewJob(scheduleConfig)
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

func convertSchedules(configs []*config.ScheduleConfig) []schedule.JobConfig {
	sc := make([]schedule.JobConfig, len(configs))
	for index, item := range configs {
		sc[index] = item
	}
	return sc
}

func removeJobs(schedulerType schedule.SchedulerConfig, profileName string, configs []schedule.JobConfig) error {
	scheduler := schedule.NewScheduler(schedulerType, profileName)
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
					clog.Warningf("service %s/%s not found", scheduleConfig.Title(), scheduleConfig.SubTitle())
				}
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

func statusJobs(schedulerType schedule.SchedulerConfig, profileName string, configs []schedule.JobConfig) error {
	scheduler := schedule.NewScheduler(schedulerType, profileName)
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
	scheduler.DisplayStatus()
	return nil
}

func getResticCommand(profileCommand string) string {
	if profileCommand == constants.SectionConfigurationRetention {
		return constants.CommandForget
	}
	return profileCommand
}
