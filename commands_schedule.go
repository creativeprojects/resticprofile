package main

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"golang.org/x/exp/maps"
)

// createSchedule accepts one argument from the commandline: --no-start
func createSchedule(_ io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	type profileJobs struct {
		scheduler schedule.SchedulerConfig
		name      string
		jobs      []*config.Schedule
	}

	allJobs := make([]profileJobs, 0, 1)

	// Step 1: Collect all jobs of all selected profiles
	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, jobs, err := getScheduleJobs(c, profileFlags)
		if err == nil {
			err = requireScheduleJobs(jobs, profileFlags)

			// Skip profile with no schedules when "--all" option is set.
			if err != nil && slices.Contains(args, "--all") {
				continue
			}
		}
		if err != nil {
			return err
		}

		// add the no-start flag to all the jobs
		if slices.Contains(args, "--no-start") {
			for id := range jobs {
				jobs[id].SetFlag("no-start", "")
			}
		}

		allJobs = append(allJobs, profileJobs{scheduler: scheduler, name: profileName, jobs: jobs})
	}

	// Step 2: Schedule all collected jobs
	for _, j := range allJobs {
		err := scheduleJobs(schedule.NewHandler(j.scheduler), j.name, j.jobs)
		if err != nil {
			return retryElevated(err, flags)
		}
	}

	return nil
}

func removeSchedule(_ io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	// Unschedule all jobs of all selected profiles
	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, jobs, err := getRemovableScheduleJobs(c, profileFlags)
		if err != nil {
			return err
		}

		err = removeJobs(schedule.NewHandler(scheduler), profileName, jobs)
		if err != nil {
			return retryElevated(err, flags)
		}
	}

	return nil
}

func statusSchedule(w io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	if !slices.Contains(args, "--all") {
		scheduler, schedules, err := getScheduleJobs(c, flags)
		if err != nil {
			return err
		}
		if len(schedules) == 0 {
			clog.Warningf("profile or group %s has no schedule", flags.name)
			return nil
		}
		statusScheduleProfileOrGroup(scheduler, schedules, flags)
	}

	for _, profileName := range selectProfiles(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)
		scheduler, schedules, err := getScheduleJobs(c, profileFlags)
		if err != nil {
			return err
		}
		// it's all fine if this profile has no schedule
		if len(schedules) == 0 {
			continue
		}
		clog.Infof("Profile/Group %q:", profileName)
		err = statusScheduleProfileOrGroup(scheduler, schedules, profileFlags)
		if err != nil {
			// display the error but keep going with the other profiles
			clog.Error(err)
		}
	}
	return nil
}

func statusScheduleProfileOrGroup(scheduler schedule.SchedulerConfig, schedules []*config.Schedule, flags commandLineFlags) error {
	err := statusJobs(schedule.NewHandler(scheduler), flags.name, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func getScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, []*config.Schedule, error) {
	global, err := c.GetGlobalSection()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot load global section: %w", err)
	}

	if c.HasProfile(flags.name) {
		profile, schedules, err := getProfileScheduleJobs(c, flags)
		if err != nil {
			return nil, nil, err
		}
		displayDeprecationNotices(profile)
		return schedule.NewSchedulerConfig(global), schedules, nil

	} else if c.HasProfileGroup(flags.name) {
		_, schedules, err := getGroupScheduleJobs(c, flags)
		if err != nil {
			return nil, nil, err
		}
		return schedule.NewSchedulerConfig(global), schedules, nil

	} else {
		return nil, nil, fmt.Errorf("profile or group %s not found", flags.name)
	}
}

func getProfileScheduleJobs(c *config.Config, flags commandLineFlags) (*config.Profile, []*config.Schedule, error) {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, fmt.Errorf("profile '%s' not found", flags.name)
		}
		return nil, nil, fmt.Errorf("cannot load profile '%s': %w", flags.name, err)
	}

	return profile, maps.Values(profile.Schedules()), nil
}

func getGroupScheduleJobs(c *config.Config, flags commandLineFlags) (*config.Group, []*config.Schedule, error) {
	group, err := c.GetProfileGroup(flags.name)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, fmt.Errorf("group '%s' not found", flags.name)
		}
		return nil, nil, fmt.Errorf("cannot load group '%s': %w", flags.name, err)
	}

	return group, maps.Values(group.Schedules()), nil
}

func requireScheduleJobs(schedules []*config.Schedule, flags commandLineFlags) error {
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", flags.name)
	}
	return nil
}

func getRemovableScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, []*config.Schedule, error) {
	scheduler, schedules, err := getScheduleJobs(c, flags)
	if err != nil {
		return nil, nil, err
	}

	// Add all undeclared schedules as remove-only configs
	// FIXME!
	// for _, command := range profile.SchedulableCommands() {
	// 	declared := false
	// 	for _, s := range schedules {
	// 		if declared = s.ScheduleOrigin().Command == command; declared {
	// 			break
	// 		}
	// 	}
	// 	if !declared {
	// 		origin := config.ScheduleOrigin(flags.name, command)
	// 		schedules = append(schedules, config.NewDefaultSchedule(c, origin))
	// 	}
	// }

	return scheduler, schedules, nil
}

func preRunSchedule(ctx *Context) error {
	if len(ctx.request.arguments) < 1 {
		return errors.New("run-schedule command expects one argument: schedule name")
	}
	scheduleName := ctx.request.arguments[0]
	// temporarily allow v2 configuration to run v1 schedules
	// if ctx.config.GetVersion() < config.Version02
	{
		commandName, profileName, ok := strings.Cut(scheduleName, "@")
		if !ok {
			return errors.New("the expected format of the schedule name is <command>@<profile-or-group-name>")
		}
		ctx.request.profile = profileName
		ctx.request.schedule = scheduleName
		ctx.command = commandName
		// remove the parameter from the arguments
		ctx.request.arguments = ctx.request.arguments[1:]

		// don't save the profile in the context now, it's only loaded but not prepared
		profile, err := ctx.config.GetProfile(profileName)
		if err != nil || profile == nil {
			return fmt.Errorf("cannot load profile '%s': %w", profileName, err)
		}
		// get the list of all scheduled commands to find the current command
		if ctx.schedule, ok = profile.Schedules()[ctx.command]; ok {
			prepareScheduledProfile(ctx)
		}
	}
	return nil
}

func prepareScheduledProfile(ctx *Context) {
	clog.Debugf("preparing scheduled profile %q", ctx.request.schedule)
	s := ctx.schedule
	// log file
	if len(s.Log) > 0 {
		ctx.logTarget = s.Log
	}
	if len(s.CommandOutput) > 0 {
		ctx.commandOutput = s.CommandOutput
	}
	// battery
	if s.IgnoreOnBatteryLessThan > 0 && !s.IgnoreOnBattery.IsStrictlyFalse() {
		ctx.stopOnBattery = s.IgnoreOnBatteryLessThan
	} else if s.IgnoreOnBattery.IsTrue() {
		ctx.stopOnBattery = 100
	}
	// lock
	if s.GetLockMode() == config.ScheduleLockModeDefault {
		if duration := s.GetLockWait(); duration > 0 {
			ctx.lockWait = duration
		}
	} else if s.GetLockMode() == config.ScheduleLockModeIgnore {
		ctx.noLock = true
	}
}

func runSchedule(_ io.Writer, cmdCtx commandContext) error {
	err := startProfileOrGroup(&cmdCtx.Context, runProfile)
	if err != nil {
		return err
	}
	return nil
}
