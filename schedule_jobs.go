package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"github.com/creativeprojects/resticprofile/util"
)

func scheduleJobs(handler schedule.Handler, configs []*config.Schedule) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	binary, err := util.Executable()
	if err != nil {
		return err
	}

	err = handler.Init()
	if err != nil {
		return err
	}
	defer handler.Close()

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

		job := schedule.NewJob(handler, scheduleConfig)
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

func removeJobs(handler schedule.Handler, configs []*config.Schedule) error {
	err := handler.Init()
	if err != nil {
		return err
	}
	defer handler.Close()

	for _, cfg := range configs {
		scheduleConfig := scheduleToConfig(cfg)
		job := schedule.NewJob(handler, scheduleConfig)

		// Skip over non-accessible, RemoveOnly jobs since they may not exist and must not causes errors
		if job.RemoveOnly() && !job.Accessible() {
			continue
		}

		// Try to remove the job
		err := job.Remove()
		if err != nil {
			if errors.Is(err, schedule.ErrScheduledJobNotFound) {
				// Display a warning and keep going. Skip message for RemoveOnly jobs since they may not exist
				if !job.RemoveOnly() {
					clog.Warningf("scheduled job %s/%s not found", scheduleConfig.ProfileName, scheduleConfig.CommandName)
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

func removeScheduledJobs(handler schedule.Handler, configFile, profileName string) error {
	err := handler.Init()
	if err != nil {
		return err
	}
	defer handler.Close()

	clog.Debugf("looking up schedules from configuration file %s", configFile)
	configs, err := handler.Scheduled(profileName)
	hasNoEntries := len(configs) == 0
	if err != nil {
		if hasNoEntries {
			return err
		}
		clog.Errorf("some configurations failed to load:\n%v", err)
	}
	if hasNoEntries {
		clog.Info("no scheduled jobs found")
		return nil
	}

	var errs error
	for _, cfg := range configs {
		if cfg.ConfigFile != configFile {
			clog.Debugf("skipping job %s/%s from configuration file %s", cfg.ProfileName, cfg.CommandName, cfg.ConfigFile)
			continue
		}
		job := schedule.NewJob(handler, &cfg)
		err = job.Remove()
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("%s/%s: %w", cfg.ProfileName, cfg.CommandName, err))
			continue
		}
		clog.Infof("scheduled job %s/%s removed", cfg.ProfileName, cfg.CommandName)
	}
	if errs != nil {
		return fmt.Errorf("failed to remove some jobs: %w", errs)
	}
	return nil
}

func statusJobs(handler schedule.Handler, profileName string, configs []*config.Schedule) error {
	err := handler.Init()
	if err != nil {
		return err
	}
	defer handler.Close()

	for _, cfg := range configs {
		scheduleConfig := scheduleToConfig(cfg)
		job := schedule.NewJob(handler, scheduleConfig)
		err := job.Status()
		if err != nil {
			if errors.Is(err, schedule.ErrScheduledJobNotFound) {
				// Display a warning and keep going
				clog.Warningf("scheduled job %s/%s not found", scheduleConfig.ProfileName, scheduleConfig.CommandName)
				continue
			}
			if errors.Is(err, schedule.ErrScheduledJobNotRunning) {
				// Display a warning and keep going
				clog.Warningf("scheduled job %s/%s is not running", scheduleConfig.ProfileName, scheduleConfig.CommandName)
				continue
			}
			return fmt.Errorf("error querying status of job %s/%s: %w",
				scheduleConfig.ProfileName,
				scheduleConfig.CommandName,
				err)
		}
	}
	err = handler.DisplayStatus(profileName)
	if err != nil {
		clog.Error(err)
	}
	return nil
}

func statusScheduledJobs(handler schedule.Handler, configFile, profileName string) error {
	err := handler.Init()
	if err != nil {
		return err
	}
	defer handler.Close()

	clog.Debugf("looking up schedules from configuration file %s", configFile)
	configs, err := handler.Scheduled(profileName)
	hasNoEntries := len(configs) == 0
	if err != nil {
		if hasNoEntries {
			return err
		}
		clog.Errorf("some configurations failed to load:\n%v", err)
	}
	if hasNoEntries {
		clog.Info("no scheduled jobs found")
		return nil
	}

	var errs error
	for _, cfg := range configs {
		if cfg.ConfigFile != configFile {
			clog.Debugf("skipping job %s/%s from configuration file %s", cfg.ProfileName, cfg.CommandName, cfg.ConfigFile)
			continue
		}
		job := schedule.NewJob(handler, &cfg)
		err := job.Status()
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to get status for job %s/%s: %w", cfg.ProfileName, cfg.CommandName, err))
		}
	}
	err = handler.DisplayStatus(profileName)
	if err != nil {
		clog.Error(err)
		errs = errors.Join(errs, fmt.Errorf("failed to display profile status: %w", err))
	}

	if errs != nil {
		return fmt.Errorf("errors on profile %s: %w", profileName, errs)
	}
	return nil
}

func scheduleToConfig(sched *config.Schedule) *schedule.Config {
	origin := sched.ScheduleOrigin()
	if !sched.HasSchedules() {
		// there's no schedule defined, so this record is for removal only
		return schedule.NewRemoveOnlyConfig(origin.Name, origin.Command)
	}
	return &schedule.Config{
		ProfileName:        origin.Name,
		CommandName:        origin.Command,
		Schedules:          sched.Schedules,
		Permission:         sched.Permission,
		RunLevel:           sched.RunLevel,
		WorkingDirectory:   "",
		Command:            "",
		Arguments:          schedule.NewCommandArguments(nil),
		Environment:        sched.Environment,
		JobDescription:     "",
		TimerDescription:   "",
		Priority:           sched.Priority,
		ConfigFile:         sched.ConfigFile,
		Flags:              sched.Flags,
		AfterNetworkOnline: sched.AfterNetworkOnline.IsTrue(),
		SystemdDropInFiles: sched.SystemdDropInFiles,
		HideWindow:         sched.HideWindow.IsTrue(),
		StartWhenAvailable: sched.StartWhenAvailable.IsTrue(),
	}
}
