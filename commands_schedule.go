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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	legacyFlagWarning = "the --legacy flag is only temporary and will be removed in version 1.0."
)

// createSchedule accepts one argument from the commandline: --no-start
func createSchedule(_ io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	type profileJobs struct {
		schedulerConfig schedule.SchedulerConfig
		name            string
		jobs            []*config.Schedule
	}

	allJobs := make([]profileJobs, 0, 1)

	// Step 1: Collect all jobs of all selected profiles
	for _, profileName := range selectProfilesAndGroups(c, flags, args) {
		profileFlags := flagsForProfile(flags, profileName)

		scheduler, jobs, _, err := getScheduleJobs(c, profileFlags)
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

		allJobs = append(allJobs, profileJobs{schedulerConfig: scheduler, name: profileName, jobs: jobs})
	}

	// Step 2: Schedule all collected jobs
	for _, j := range allJobs {
		err := scheduleJobs(schedule.NewHandler(j.schedulerConfig), j.jobs)
		if err != nil {
			return retryElevated(err, flags)
		}
	}

	return nil
}

func removeSchedule(_ io.Writer, ctx commandContext) error {
	var err error
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	if slices.Contains(args, "--legacy") {
		clog.Warning(legacyFlagWarning)
		// Unschedule all jobs of all selected profiles
		for _, profileName := range selectProfilesAndGroups(c, flags, args) {
			profileFlags := flagsForProfile(flags, profileName)

			schedulerConfig, jobs, err := getRemovableScheduleJobs(c, profileFlags)
			if err != nil {
				return err
			}

			err = removeJobs(schedule.NewHandler(schedulerConfig), jobs)
			if err != nil {
				err = retryElevated(err, flags)
			}
			if err != nil {
				// we keep trying to remove the other jobs
				clog.Error(err)
			}
		}
		return nil
	}

	profileName := ctx.request.profile
	if slices.Contains(args, "--all") {
		// Unschedule all jobs of all profiles
		profileName = ""
	}
	schedulerConfig := schedule.NewSchedulerConfig(ctx.global)
	err = removeScheduledJobs(schedule.NewHandler(schedulerConfig), ctx.config.GetConfigFile(), profileName)
	if err != nil {
		err = retryElevated(err, flags)
	}
	return err
}

func statusSchedule(w io.Writer, ctx commandContext) error {
	c := ctx.config
	flags := ctx.flags
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	if slices.Contains(flags.resticArgs, "--legacy") {
		clog.Warning(legacyFlagWarning)
		// single profile or group
		if !slices.Contains(args, "--all") {
			schedulerConfig, schedules, _, err := getScheduleJobs(c, flags)
			if err != nil {
				return err
			}
			if len(schedules) == 0 {
				clog.Warningf("profile or group %s has no schedule", flags.name)
				return nil
			}
			err = statusScheduleProfileOrGroup(schedulerConfig, schedules, flags)
			if err != nil {
				return err
			}
			return nil
		}

		// all profiles and groups
		for _, profileName := range selectProfilesAndGroups(c, flags, args) {
			profileFlags := flagsForProfile(flags, profileName)
			scheduler, schedules, schedulable, err := getScheduleJobs(c, profileFlags)
			if err != nil {
				return err
			}
			// it's all fine if this profile has no schedule
			if len(schedules) == 0 {
				continue
			}
			clog.Infof("%s %q:", cases.Title(language.English).String(schedulable.Kind()), profileName)
			err = statusScheduleProfileOrGroup(scheduler, schedules, profileFlags)
			if err != nil {
				// display the error but keep going with the other profiles
				clog.Error(err)
			}
		}
	}
	profileName := ctx.request.profile
	if slices.Contains(args, "--all") {
		// Unschedule all jobs of all profiles
		profileName = ""
	}
	schedulerConfig := schedule.NewSchedulerConfig(ctx.global)
	err := statusScheduledJobs(schedule.NewHandler(schedulerConfig), ctx.config.GetConfigFile(), profileName)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

// selectProfilesAndGroups returns a list with length >= 1, containing profile and group names that have been selected in flags or extra args.
// With "--all" set in args, names of all profiles and groups are returned, otherwise flags.name is returned as-is.
func selectProfilesAndGroups(c *config.Config, flags commandLineFlags, args []string) []string {
	schedulables := make([]string, 0, 1)

	// Check for --all or groups
	if slices.Contains(args, "--all") {
		schedulables = append(schedulables, c.GetProfileNames()...)
		schedulables = append(schedulables, c.GetGroupNames()...)
	}

	// Fallback add profile name from flags
	if len(schedulables) == 0 {
		schedulables = append(schedulables, flags.name)
	}

	return schedulables
}

// flagsForProfile returns a copy of flags with name set to profileName.
func flagsForProfile(flags commandLineFlags, profileName string) commandLineFlags {
	flags.name = profileName
	return flags
}

func statusScheduleProfileOrGroup(schedulerConfig schedule.SchedulerConfig, schedules []*config.Schedule, flags commandLineFlags) error {
	err := statusJobs(schedule.NewHandler(schedulerConfig), flags.name, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func getScheduleJobs(c *config.Config, flags commandLineFlags) (schedule.SchedulerConfig, []*config.Schedule, config.Schedulable, error) {
	global, err := c.GetGlobalSection()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot load global section: %w", err)
	}

	if c.HasProfile(flags.name) {
		profile, schedules, err := getProfileScheduleJobs(c, flags)
		if err != nil {
			return nil, nil, nil, err
		}
		displayDeprecationNotices(profile)
		return schedule.NewSchedulerConfig(global), schedules, profile, nil

	} else if c.HasProfileGroup(flags.name) {
		group, schedules, err := getGroupScheduleJobs(c, flags)
		if err != nil {
			return nil, nil, nil, err
		}
		return schedule.NewSchedulerConfig(global), schedules, group, nil

	} else {
		return nil, nil, nil, fmt.Errorf("profile or group '%s': %w", flags.name, config.ErrNotFound)
	}
}

func getProfileScheduleJobs(c *config.Config, flags commandLineFlags) (*config.Profile, []*config.Schedule, error) {
	profile, err := c.GetProfile(flags.name)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, fmt.Errorf("profile '%s': %w", flags.name, err)
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
	scheduler, schedules, schedulable, err := getScheduleJobs(c, flags)
	if err != nil {
		return nil, nil, err
	}

	// Add all undeclared schedules as remove-only configs
	for _, command := range schedulable.SchedulableCommands() {
		declared := false
		for _, s := range schedules {
			if declared = s.ScheduleOrigin().Command == command; declared {
				break
			}
		}
		if !declared {
			origin := config.ScheduleOrigin(flags.name, command)
			schedules = append(schedules, config.NewDefaultSchedule(c, origin))
		}
	}

	return scheduler, schedules, nil
}

func preRunSchedule(ctx *Context) error {
	if len(ctx.request.arguments) < 1 {
		return errors.New("run-schedule command expects one argument: schedule name")
	}
	scheduleName := ctx.request.arguments[0]
	commandName, profileName, ok := strings.Cut(scheduleName, "@")
	if !ok {
		return errors.New("the expected format of the schedule name is <command>@<profile-or-group-name>")
	}
	ctx.request.profile = profileName
	ctx.request.schedule = scheduleName
	ctx.command = commandName
	// remove the parameter from the arguments
	ctx.request.arguments = ctx.request.arguments[1:]

	var schedulable config.Schedulable
	if ctx.config.HasProfile(profileName) {
		// don't save the profile in the context now, it's only loaded but not prepared
		profile, err := ctx.config.GetProfile(profileName)
		if err != nil || profile == nil {
			return fmt.Errorf("cannot load profile '%s': %w", profileName, err)
		}
		schedulable = profile
	} else if ctx.config.HasProfileGroup(profileName) {
		group, err := ctx.config.GetProfileGroup(profileName)
		if err != nil || group == nil {
			return fmt.Errorf("cannot load group '%s': %w", profileName, err)
		}
		schedulable = group
	} else {
		return fmt.Errorf("profile or group %q: %w", profileName, config.ErrNotFound)
	}
	// get the list of all scheduled commands to find the current command
	if ctx.schedule, ok = schedulable.Schedules()[ctx.command]; ok {
		clog.Debugf("preparing scheduled %s %q", schedulable.Kind(), ctx.request.schedule)
		prepareScheduledProfile(ctx)
	}
	return nil
}

func prepareScheduledProfile(ctx *Context) {
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
