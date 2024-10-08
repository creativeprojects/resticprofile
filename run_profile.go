package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor/prom"
	"github.com/creativeprojects/resticprofile/monitor/status"
)

// startProfileOrGroup starts a profile or a group of profiles based on the provided context.
// It first checks if the requested profile exists and runs it. If the profile is part of a group,
// it runs all profiles in the group sequentially. If any profile in the group fails and the
// ContinueOnError flag is set, it continues with the next profile. Otherwise, it stops and returns the error.
//
// Parameters:
//   - ctx: A pointer to the Context struct containing configuration and request details.
//   - runProfile: A function that takes a Context pointer and returns an error.
//
// Returns:
//   - error: An error if the profile or group run fails, or if the requested profile is not found.
func startProfileOrGroup(ctx *Context, runProfile func(ctx *Context) error) error {
	if ctx.config.HasProfile(ctx.request.profile) {
		// if running as a systemd timer
		notifyStart()
		defer notifyStop()

		// Single profile run
		err := runProfile(ctx)
		if err != nil {
			return err
		}

	} else if ctx.config.HasProfileGroup(ctx.request.profile) {
		// Group run
		group, err := ctx.config.GetProfileGroup(ctx.request.profile)
		if err != nil {
			clog.Errorf("cannot load group '%s': %v", ctx.request.profile, err)
		}
		if group != nil && len(group.Profiles) > 0 {
			// if running as a systemd timer
			notifyStart()
			defer notifyStop()

			// profile name is the group name
			groupName := ctx.request.profile

			for i, profileName := range group.Profiles {
				clog.Debugf("[%d/%d] starting profile '%s' from group '%s'", i+1, len(group.Profiles), profileName, groupName)
				ctx = ctx.WithProfile(profileName).WithGroup(groupName)
				err = runProfile(ctx)
				if err != nil {
					if group.ContinueOnError.IsTrue() || (ctx.global.GroupContinueOnError && group.ContinueOnError.IsUndefined()) {
						// keep going to the next profile
						clog.Error(err)
						continue
					}
					// fail otherwise
					return err
				}
			}
		}

	} else {
		return fmt.Errorf("%w: %q", ErrProfileNotFound, ctx.request.profile)
	}
	return nil
}

func openProfile(c *config.Config, profileName string) (profile *config.Profile, cleanup func(), err error) {
	done := false
	for attempts := 3; attempts > 0 && !done; attempts-- {
		profile, err = c.GetProfile(profileName)
		if err != nil || profile == nil {
			err = fmt.Errorf("cannot load profile '%s': %w", profileName, err)
			break
		}

		done = true

		// Adjust baseDir if needed
		if len(profile.BaseDir) > 0 {
			var currentDir string
			currentDir, err = os.Getwd()
			if err != nil {
				err = fmt.Errorf("changing base directory not allowed as current directory is unknown in profile %q: %w", profileName, err)
				break
			}

			if baseDir, _ := filepath.Abs(profile.BaseDir); filepath.ToSlash(baseDir) != filepath.ToSlash(currentDir) {
				if cleanup == nil {
					cleanup = func() {
						if e := os.Chdir(currentDir); e != nil {
							panic(fmt.Errorf(`fatal: failed restoring working directory "%s": %w`, currentDir, e))
						}
					}
				}

				if err = os.Chdir(baseDir); err == nil {
					clog.Infof("profile '%s': base directory is %q", profileName, baseDir)
					done = false // reload the profile as .CurrentDir & .Env has changed
				} else {
					err = fmt.Errorf(`cannot change to base directory "%s" in profile %q: %w`, baseDir, profileName, err)
					break
				}
			}
		}
	}

	if cleanup == nil {
		cleanup = func() {
			// nothing to do
		}
	}
	return
}

func runProfile(ctx *Context) error {
	profile, cleanup, err := openProfile(ctx.config, ctx.request.profile)
	defer cleanup()
	if err != nil {
		return err
	}
	ctx.profile = profile

	displayDeprecationNotices(profile)
	ctx.config.DisplayConfigurationIssues()

	// Send the quiet/verbose down to restic as well (override profile configuration)
	if ctx.flags.quiet {
		profile.Quiet = true
		profile.Verbose = constants.VerbosityNone
	}
	if ctx.flags.verbose {
		profile.Verbose = constants.VerbosityLevel1
		profile.Quiet = false
	}
	if ctx.flags.veryVerbose {
		profile.Verbose = constants.VerbosityLevel3
		profile.Quiet = false
	}

	// change log filter according to profile settings
	if profile.Quiet {
		changeLevelFilter(clog.LevelWarning)
	} else if profile.Verbose > constants.VerbosityNone && !ctx.flags.veryVerbose {
		changeLevelFilter(clog.LevelDebug)
	}

	// tell the profile what version of restic is in use
	if e := profile.SetResticVersion(ctx.global.ResticVersion); e != nil {
		clog.Warningf("restic version %q is no valid semver: %s", ctx.global.ResticVersion, e.Error())
	}

	// Specific case for the "host" flag where an empty value should be replaced by the hostname
	hostname := "none"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}
	profile.SetHost(hostname)

	if ctx.request.schedule != "" {
		// this is a scheduled profile
		loadScheduledProfile(ctx)
	}

	// Catch CTR-C keypress, or other signal sent by a service manager (systemd)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	// remove signal catch before leaving
	defer signal.Stop(sigChan)

	ctx.sigChan = sigChan
	wrapper := newResticWrapper(ctx)

	if ctx.noLock {
		wrapper.ignoreLock()
	} else if ctx.lockWait > 0 {
		wrapper.maxWaitOnLock(ctx.lockWait)
	}

	// add progress receivers if necessary
	if profile.StatusFile != "" {
		wrapper.addProgress(status.NewProgress(profile, status.NewStatus(profile.StatusFile)))
	}
	if profile.PrometheusPush != "" || profile.PrometheusSaveToFile != "" {
		wrapper.addProgress(prom.NewProgress(profile, prom.NewMetrics(profile.Name, ctx.request.group, version, profile.PrometheusLabels)))
	}

	err = wrapper.runProfile()
	if err != nil {
		return err
	}
	return nil
}

func loadScheduledProfile(ctx *Context) {
	ctx.schedule = ctx.profile.Schedules()[ctx.command]
}
