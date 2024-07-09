package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"text/tabwriter"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/monitor/prom"
	"github.com/creativeprojects/resticprofile/monitor/status"
	"github.com/creativeprojects/resticprofile/preventsleep"
	"github.com/creativeprojects/resticprofile/priority"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/shutdown"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/spf13/pflag"
)

// These fields are populated by the goreleaser build
var (
	version = "0.28.0-dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	var exitCode = 0
	var err error

	// trick to run all defer functions before returning with an exit code
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	// run shutdown hooks just before returning an exit code
	defer shutdown.RunHooks()

	args := os.Args[1:]
	_, flags, flagErr := loadFlags(args)
	if flagErr != nil && flagErr != pflag.ErrHelp {
		term.Println(flagErr)
		_ = displayHelpCommand(os.Stdout, commandContext{
			ownCommands: ownCommands,
			Context: Context{
				flags: flags,
				request: Request{
					arguments: args,
				},
			},
		})
		exitCode = 2
		return
	}

	if flags.wait {
		// keep the console running at the end of the program
		// so we can see what's going on
		defer func() {
			term.Println("\n\nPress the Enter Key to continue...")
			fmt.Scanln()
		}()
	}

	if flags.isChild {
		if flags.parentPort == 0 {
			exitCode = 10
			return
		}
	}

	// help
	if flags.help || errors.Is(flagErr, pflag.ErrHelp) {
		_ = displayHelpCommand(os.Stdout, commandContext{
			ownCommands: ownCommands,
			Context: Context{
				flags: flags,
				request: Request{
					arguments: args,
				},
			},
		})
		return
	}

	// logger setup logic - is delayed until config was loaded (or attempted)
	setupLogging := func(ctx *Context) (logCloser func()) {
		logCloser = func() {}

		if flags.isChild {
			// use a remote logger
			client := remote.NewClient(flags.parentPort)
			logCloser = func() { _ = client.Done() }
			setupRemoteLogger(flags, client)

			// also redirect the terminal through the client
			term.SetAllOutput(term.NewRemoteTerm(client))
		} else {
			logTarget, commandOutput := "", ""
			if ctx != nil {
				logTarget = ctx.logTarget
				commandOutput = ctx.commandOutput
				if ctx.request.command == constants.CommandCat ||
					ctx.request.command == constants.CommandDump {
					clog.Debugf("redirecting console to stderr for command %q", ctx.request.command)
					flags.stderr = true
				}
				term.PrintToError = flags.stderr
			}
			if logTarget != "" && logTarget != "-" {
				if closer, err := setupTargetLogger(flags, logTarget, commandOutput); err == nil {
					logCloser = func() { _ = closer.Close() }
				} else {
					// fallback to a console logger
					setupConsoleLogger(flags)
					clog.Errorf("cannot open log target: %s", err)
				}
			} else {
				// use the console logger
				setupConsoleLogger(flags)
			}
		}
		return
	}

	// keep this one last if possible (so it will be first at the end)
	defer showPanicData()

	banner()

	// resticprofile own commands (configuration file may not be loaded)
	if len(flags.resticArgs) > 0 {
		if ownCommands.Exists(flags.resticArgs[0], false) {
			ctx := &Context{
				flags: flags,
				request: Request{
					command:   flags.resticArgs[0],
					arguments: flags.resticArgs[1:],
				},
			}
			// try to load the config and setup logging for own command
			cfg, global, err := loadConfig(flags, true)
			if err == nil {
				ctx = ctx.WithConfig(cfg, global)
			}
			closeLogger := setupLogging(ctx)
			defer closeLogger()
			err = ownCommands.Run(ctx)
			if err != nil {
				clog.Error(err)
				exitCode = 1
				var ownCommandError *ownCommandError
				if errors.As(err, &ownCommandError) {
					exitCode = ownCommandError.ExitCode()
				}
				return
			}
			return
		}
	}

	// Load the now mandatory configuration and setup logging (before returning an error)
	ctx, err := loadContext(flags)
	closeLogger := setupLogging(ctx)
	defer closeLogger()
	if err != nil {
		clog.Error(err)
		exitCode = 1
		return
	}

	// check if we're running on battery
	if shouldStopOnBattery(ctx.stopOnBattery) {
		exitCode = 3
		return
	}

	// prevent computer from sleeping
	var caffeinate *preventsleep.Caffeinate
	if ctx.global.PreventSleep {
		clog.Debug("preventing the system from sleeping")
		caffeinate = preventsleep.New()
		err = caffeinate.Start()
		if err != nil {
			clog.Errorf("preventing system sleep: %s", err)
		}
	}
	// and stop at the end
	defer func() {
		if caffeinate != nil {
			err = caffeinate.Stop()
			if err != nil {
				clog.Error(err)
			}
		}
	}()

	// Check memory pressure
	if ctx.global.MinMemory > 0 {
		avail := free()
		if avail > 0 && avail < ctx.global.MinMemory {
			clog.Errorf("available memory is < %v MB (option 'min-memory' in the 'global' section)", ctx.global.MinMemory)
			exitCode = 1
			return
		}
	}

	if !flags.noPriority {
		err = setPriority(ctx.global.Nice, ctx.global.Priority)
		if err != nil {
			clog.Warning(err)
		}

		if ctx.global.IONice {
			err = priority.SetIONice(ctx.global.IONiceClass, ctx.global.IONiceLevel)
			if err != nil {
				clog.Warning(err)
			}
		}
	}

	resticBinary, err := detectResticBinary(ctx.global)
	if err != nil {
		clog.Error(err)
		clog.Warning("you can specify the path of the restic binary in the global section of the configuration file (restic-binary)")
		exitCode = 1
		return
	}
	ctx = ctx.WithBinary(resticBinary)

	// resticprofile own commands (with configuration file)
	if ownCommands.Exists(ctx.request.command, true) {
		err = ownCommands.Run(ctx)
		if err != nil {
			clog.Error(err)
			exitCode = 1
			var ownCommandError *ownCommandError
			if errors.As(err, &ownCommandError) {
				exitCode = ownCommandError.ExitCode()
			}
			return
		}
		return
	}

	// since it's not a resticprofile command, it's a restic command
	ctx = ctx.WithCommand(ctx.request.command)

	// it wasn't an internal command so we run a profile
	err = startProfileOrGroup(ctx)
	if err != nil {
		clog.Error(err)
		if errors.Is(err, ErrProfileNotFound) {
			displayProfiles(os.Stdout, ctx.config, flags)
			displayGroups(os.Stdout, ctx.config, flags)
		}
		exitCode = 1
		return
	}
}

func banner() {
	clog.Debugf("resticprofile %s compiled with %s", version, runtime.Version())
}

func loadConfig(flags commandLineFlags, silent bool) (cfg *config.Config, global *config.Global, err error) {
	var configFile string
	if configFile, err = filesearch.FindConfigurationFile(flags.config); err == nil {
		if configFile != flags.config && !silent {
			clog.Infof("using configuration file: %s", configFile)
		}

		if cfg, err = config.LoadFile(configFile, flags.format); err == nil {
			global, err = cfg.GetGlobalSection()
			if err != nil {
				err = fmt.Errorf("cannot load global configuration: %w", err)
			}
		} else {
			err = fmt.Errorf("cannot load configuration file: %w", err)
		}
	}
	return
}

// loadContext loads the configuration and creates a context.
func loadContext(flags commandLineFlags) (*Context, error) {
	cfg, global, err := loadConfig(flags, false)
	if err != nil {
		return nil, err
	}
	return CreateContext(flags, global, cfg, ownCommands)
}

func setPriority(nice int, class string) error {
	var err error

	if class != "" {
		if classID, ok := constants.PriorityValues[class]; ok {
			err = priority.SetClass(classID)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("incorrect value '%s' for priority in global section", class)
		}
		return nil
	}
	if nice != 0 {
		err = priority.SetNice(nice)
		if err != nil {
			return err
		}
	}
	return nil
}

func shouldStopOnBattery(batteryLimit int) bool {
	if batteryLimit > 0 && batteryLimit <= constants.BatteryFull {
		battery, charge, err := IsRunningOnBattery()
		if err != nil {
			clog.Errorf("cannot check if the computer is running on battery: %s", err)
		}
		if battery {
			if batteryLimit == constants.BatteryFull {
				clog.Warning("running on battery, leaving now")
				return true
			}
			if charge < batteryLimit {
				clog.Warningf("running on battery (%d%%), leaving now", charge)
				return true
			}
			clog.Infof("running on battery with enough charge (%d%%)", charge)
		}
	}
	return false
}

func detectResticBinary(global *config.Global) (string, error) {
	resticBinary, err := filesearch.FindResticBinary(global.ResticBinary)
	if err != nil {
		return "", fmt.Errorf("cannot find restic: %w", err)
	}
	// detect restic version
	if len(global.ResticVersion) == 0 {
		if global.ResticVersion, err = restic.GetVersion(resticBinary); err != nil {
			clog.Warningf("assuming restic is at latest known version ; %s", err.Error())
			global.ResticVersion = restic.AnyVersion
		}
	}
	if len(global.ResticVersion) > 0 {
		clog.Debugf("using restic %s", global.ResticVersion)
	}
	return resticBinary, nil
}

func startProfileOrGroup(ctx *Context) error {
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
					if ctx.global.GroupContinueOnError && group.ContinueOnError.IsTrueOrUndefined() {
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

	displayProfileDeprecationNotices(profile)
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

func loadScheduledProfile(ctx *Context) error {
	ctx.schedule, _ = ctx.profile.Schedules()[ctx.command]
	return nil
}

// randomBool returns true for Heads and false for Tails
func randomBool() bool {
	return rand.Int31n(10000) < 5000
}

func free() uint64 {
	const oneMB = 1048576
	mem, err := memory.Get()
	if err != nil {
		clog.Info("OS memory information not available")
		return 0
	}
	avail := (mem.Total - mem.Used) / oneMB
	clog.Debugf("memory available: %vMB", avail)
	return avail
}

func showPanicData() {
	if r := recover(); r != nil {
		message := `
===============================================================
uh-oh! resticprofile crashed miserably :-(
Can you please open an issue on github including these details:
===============================================================
`
		fmt.Fprint(os.Stderr, message)
		w := tabwriter.NewWriter(os.Stderr, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "os", runtime.GOOS)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "arch", runtime.GOARCH)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "version", version)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "commit", commit)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "compiled", date)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "by", builtBy)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "error", r)
		_, _ = fmt.Fprintf(w, "\t%s:\n%s\n", "stack", getStack(3)) // skip calls to getStack - showPanicData - panic
		w.Flush()
		fmt.Fprint(os.Stderr, "===============================================================\n")
	}
}
