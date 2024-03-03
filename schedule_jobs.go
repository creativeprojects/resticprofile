package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
)

func scheduleJobs(handler schedule.Handler, profileName string, configs []*config.Schedule) error {
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

	for _, cfg := range configs {
		scheduleConfig := scheduleToConfig(cfg)
		scheduleName := scheduleConfig.CommandName + "@" + scheduleConfig.ProfileName
		args := []string{
			"--no-ansi",
			"--config",
			scheduleConfig.ConfigFile,
			"run-schedule",
			scheduleName,
		}

		scheduleConfig.SetCommand(wd, binary, args)
		scheduleConfig.JobDescription =
			fmt.Sprintf("resticprofile %s for profile %s in %s", scheduleConfig.CommandName, scheduleConfig.ProfileName, scheduleConfig.ConfigFile)
		scheduleConfig.TimerDescription =
			fmt.Sprintf("%s timer for profile %s in %s", scheduleConfig.CommandName, scheduleConfig.ProfileName, scheduleConfig.ConfigFile)

		job := scheduler.NewJob(scheduleConfig)
		err = job.Create()
		if err != nil {
			return fmt.Errorf("error creating job %s/%s: %w",
				scheduleConfig.ProfileName,
				scheduleConfig.CommandName,
				err)
		}
		clog.Infof("scheduled job %s/%s created", scheduleConfig.ProfileName, scheduleConfig.CommandName)
	}
	return nil
}

func removeJobs(handler schedule.Handler, profileName string, configs []*config.Schedule) error {
	scheduler := schedule.NewScheduler(handler, profileName)
	err := scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, cfg := range configs {
		scheduleConfig := scheduleToConfig(cfg)
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
					clog.Warningf("service %s/%s not found", scheduleConfig.ProfileName, scheduleConfig.CommandName)
				}
				continue
			}
			return fmt.Errorf("error removing job %s/%s: %w",
				scheduleConfig.ProfileName,
				scheduleConfig.CommandName,
				err)
		}

		clog.Infof("scheduled job %s/%s removed", scheduleConfig.ProfileName, scheduleConfig.CommandName)
	}
	return nil
}

func statusJobs(handler schedule.Handler, profileName string, configs []*config.Schedule) error {
	scheduler := schedule.NewScheduler(handler, profileName)
	err := scheduler.Init()
	if err != nil {
		return err
	}
	defer scheduler.Close()

	for _, cfg := range configs {
		scheduleConfig := scheduleToConfig(cfg)
		job := scheduler.NewJob(scheduleConfig)
		err := job.Status()
		if err != nil {
			if errors.Is(err, schedule.ErrorServiceNotFound) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s not found", scheduleConfig.ProfileName, scheduleConfig.CommandName)
				continue
			}
			if errors.Is(err, schedule.ErrorServiceNotRunning) {
				// Display a warning and keep going
				clog.Warningf("service %s/%s is not running", scheduleConfig.ProfileName, scheduleConfig.CommandName)
				continue
			}
			return fmt.Errorf("error querying status of job %s/%s: %w",
				scheduleConfig.ProfileName,
				scheduleConfig.CommandName,
				err)
		}
	}
	scheduler.DisplayStatus()
	return nil
}

func scheduleToConfig(sched *config.Schedule) *schedule.Config {
	origin := sched.ScheduleOrigin()
	if !sched.HasSchedules() {
		// there's no schedule defined, so this record is for removal only
		return schedule.NewRemoveOnlyConfig(origin.Name, origin.Command)
	}
	return &schedule.Config{
		ProfileName:             origin.Name,
		CommandName:             origin.Command,
		Schedules:               sched.Schedules,
		Permission:              sched.Permission,
		WorkingDirectory:        "",
		Command:                 "",
		Arguments:               []string{},
		Environment:             sched.Environment,
		JobDescription:          "",
		TimerDescription:        "",
		Priority:                sched.Priority,
		ConfigFile:              sched.ConfigFile,
		Flags:                   sched.Flags,
		IgnoreOnBattery:         sched.IgnoreOnBattery.IsTrue(),
		IgnoreOnBatteryLessThan: sched.IgnoreOnBatteryLessThan,
	}
}
