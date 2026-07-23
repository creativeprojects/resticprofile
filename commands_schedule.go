package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/schedule"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	legacyFlagWarning = "the --legacy flag is only temporary and will be removed in version 1.0.0"
)

// createSchedule command
func createSchedule(ctx commandContext) error {
	c := ctx.config
	request := ctx.request
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	type profileJobs struct {
		schedulerConfig schedule.SchedulerConfig
		name            string
		jobs            []*config.Schedule
	}

	allJobs := make([]profileJobs, 0, 1)

	// Step 1: Collect all jobs of all selected profiles
	for _, profileName := range selectProfilesAndGroups(c, request.profile, args) {
		scheduler, jobs, _, err := getScheduleJobs(c, profileName)
		if err == nil {
			err = requireScheduleJobs(jobs, profileName)

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
		if slices.Contains(args, "--reload") {
			for id := range jobs {
				jobs[id].SetFlag("reload", "")
			}
		}

		allJobs = append(allJobs, profileJobs{schedulerConfig: scheduler, name: profileName, jobs: jobs})
	}

	// Step 2: Schedule all collected jobs
	for _, j := range allJobs {
		err := scheduleJobs(schedule.NewHandler(j.schedulerConfig), j.jobs)
		if err != nil {
			return retryElevated(err, ctx.flags)
		}
	}

	// Optional: if --metrics-port is set, keep the process alive and serve
	// the profiles' prometheus-save-to-file on /metrics. This lets the schedule
	// command double as a metrics sidecar when running under crond in a container.
	if err := blockForMetricsServer(ctx); err != nil {
		return err
	}

	return nil
}

// blockForMetricsServer starts a /metrics HTTP server for the prometheus-save-to-file
// of the selected profiles, then blocks until SIGINT/SIGTERM. Returns nil immediately
// (and the function returns) when --metrics-port is not set or no profile has a
// prometheus-save-to-file configured. A warning is logged in the latter case so the
// user knows metrics were requested but no file is available.
func blockForMetricsServer(ctx commandContext) error {
	if ctx.flags.metricsPort <= 0 {
		return nil
	}

	metricFiles := collectPrometheusSaveToFiles(ctx)
	if len(metricFiles) == 0 {
		clog.Warningf("metrics server not started: no profile has prometheus-save-to-file set")
		return nil
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	errCh := make(chan error, len(metricFiles))
	for _, file := range metricFiles {
		file := file
		go func() {
			if err := newMetricsServer(ctx.flags.metricsPort, file).run(quit); err != nil {
				clog.Errorf("metrics server failed for %q: %v", file, err)
				errCh <- err
			}
		}()
	}

	clog.Infof("schedule command will keep the process alive for the metrics server; send SIGINT/SIGTERM to stop")
	select {
	case <-quit:
		clog.Info("shutting down metrics server (schedule)")
		return nil
	case err := <-errCh:
		// best-effort: close quit so all running goroutines shut down
		close(quit)
		return err
	}
}

// collectPrometheusSaveToFiles returns the unique non-empty prometheus-save-to-file
// paths of all profiles selected by the schedule command (--all or a specific profile).
// If a group is selected, the prometheus-save-to-file of each member profile is used.
func collectPrometheusSaveToFiles(ctx commandContext) []string {
	seen := make(map[string]struct{})
	var files []string
	args := ctx.request.arguments
	add := func(name string) {
		profile, err := ctx.config.GetProfile(name)
		if err != nil || profile == nil || profile.PrometheusSaveToFile == "" {
			return
		}
		if _, ok := seen[profile.PrometheusSaveToFile]; ok {
			return
		}
		seen[profile.PrometheusSaveToFile] = struct{}{}
		files = append(files, profile.PrometheusSaveToFile)
	}

	for _, name := range selectProfilesAndGroups(ctx.config, ctx.request.profile, args) {
		if ctx.config.HasProfile(name) {
			add(name)
			continue
		}
		if group, err := ctx.config.GetProfileGroup(name); err == nil && group != nil {
			for _, member := range group.Profiles {
				add(member)
			}
		}
	}
	return files
}

func removeSchedule(ctx commandContext) error {
	var err error
	c := ctx.config
	request := ctx.request
	args := ctx.request.arguments

	if slices.Contains(args, "--legacy") {
		clog.Warning(legacyFlagWarning)
		// Unschedule all jobs of all selected profiles
		for _, profileName := range selectProfilesAndGroups(c, request.profile, args) {
			schedulerConfig, jobs, err := getRemovableScheduleJobs(c, profileName)
			if err != nil {
				return err
			}

			err = removeJobs(schedule.NewHandler(schedulerConfig), jobs)
			if err != nil {
				err = retryElevated(err, ctx.flags)
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
		return retryElevated(err, ctx.flags)
	}
	return nil
}

func statusSchedule(ctx commandContext) error {
	c := ctx.config
	request := ctx.request
	args := ctx.request.arguments

	defer c.DisplayConfigurationIssues()

	if slices.Contains(args, "--legacy") {
		clog.Warning(legacyFlagWarning)
		// single profile or group
		if !slices.Contains(args, "--all") {
			schedulerConfig, schedules, _, err := getScheduleJobs(c, request.profile)
			if err != nil {
				return err
			}
			if len(schedules) == 0 {
				clog.Warningf("profile or group %s has no schedule", request.profile)
				return nil
			}
			err = statusScheduleProfileOrGroup(schedulerConfig, schedules, ctx.flags, request.profile)
			if err != nil {
				return err
			}
			return nil
		}

		// all profiles and groups
		for _, profileName := range selectProfilesAndGroups(c, request.profile, args) {
			scheduler, schedules, schedulable, err := getScheduleJobs(c, profileName)
			if err != nil {
				return err
			}
			// it's all fine if this profile has no schedule
			if len(schedules) == 0 {
				continue
			}
			clog.Infof("%s %q:", cases.Title(language.English).String(schedulable.Kind()), profileName)
			err = statusScheduleProfileOrGroup(scheduler, schedules, ctx.flags, profileName)
			if err != nil {
				// display the error but keep going with the other profiles
				clog.Error(err)
			}
		}
	}
	profileName := ctx.request.profile
	if slices.Contains(args, "--all") {
		// display all jobs of all profiles
		profileName = ""
	}
	schedulerConfig := schedule.NewSchedulerConfig(ctx.global)
	err := statusScheduledJobs(schedule.NewHandler(schedulerConfig), ctx.config.GetConfigFile(), profileName)
	if err != nil {
		return retryElevated(err, ctx.flags)
	}
	return nil
}

// selectProfilesAndGroups returns a list with length >= 1, containing profile and group names that have been selected in flags or extra args.
// With "--all" set in args, names of all profiles and groups are returned, otherwise profileName is returned as-is.
func selectProfilesAndGroups(c *config.Config, profileName string, args []string) []string {
	schedulables := make([]string, 0, 1)

	// Check for --all or groups
	if slices.Contains(args, "--all") {
		schedulables = append(schedulables, c.GetProfileNames()...)
		schedulables = append(schedulables, c.GetGroupNames()...)
	}

	// Fallback add profile name from flags
	if len(schedulables) == 0 {
		schedulables = append(schedulables, profileName)
	}

	return schedulables
}

func statusScheduleProfileOrGroup(schedulerConfig schedule.SchedulerConfig, schedules []*config.Schedule, flags commandLineFlags, profileName string) error {
	err := statusJobs(schedule.NewHandler(schedulerConfig), profileName, schedules)
	if err != nil {
		return retryElevated(err, flags)
	}
	return nil
}

func getScheduleJobs(c *config.Config, profileName string) (schedule.SchedulerConfig, []*config.Schedule, config.Schedulable, error) {
	global, err := c.GetGlobalSection()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot load global section: %w", err)
	}

	if c.HasProfile(profileName) {
		profile, schedules, err := getProfileScheduleJobs(c, profileName)
		if err != nil {
			return nil, nil, nil, err
		}
		displayDeprecationNotices(profile)
		return schedule.NewSchedulerConfig(global), schedules, profile, nil

	} else if c.HasProfileGroup(profileName) {
		group, schedules, err := getGroupScheduleJobs(c, profileName)
		if err != nil {
			return nil, nil, nil, err
		}
		return schedule.NewSchedulerConfig(global), schedules, group, nil

	} else {
		return nil, nil, nil, fmt.Errorf("profile or group '%s': %w", profileName, config.ErrNotFound)
	}
}

func getProfileScheduleJobs(c *config.Config, profileName string) (*config.Profile, []*config.Schedule, error) {
	profile, err := c.GetProfile(profileName)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, fmt.Errorf("profile '%s': %w", profileName, err)
		}
		return nil, nil, fmt.Errorf("cannot load profile '%s': %w", profileName, err)
	}

	return profile, slices.Collect(maps.Values(profile.Schedules())), nil
}

func getGroupScheduleJobs(c *config.Config, profileName string) (*config.Group, []*config.Schedule, error) {
	group, err := c.GetProfileGroup(profileName)
	if err != nil {
		if errors.Is(err, config.ErrNotFound) {
			return nil, nil, fmt.Errorf("group '%s' not found", profileName)
		}
		return nil, nil, fmt.Errorf("cannot load group '%s': %w", profileName, err)
	}

	return group, slices.Collect(maps.Values(group.Schedules())), nil
}

func requireScheduleJobs(schedules []*config.Schedule, profileName string) error {
	if len(schedules) == 0 {
		return fmt.Errorf("no schedule found for profile '%s'", profileName)
	}
	return nil
}

func getRemovableScheduleJobs(c *config.Config, profileName string) (schedule.SchedulerConfig, []*config.Schedule, error) {
	scheduler, schedules, schedulable, err := getScheduleJobs(c, profileName)
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
			origin := config.ScheduleOrigin(profileName, command)
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

func runSchedule(cmdCtx commandContext) error {
	err := startProfileOrGroup(&cmdCtx.Context, runProfile)
	if err != nil {
		return err
	}
	return nil
}
