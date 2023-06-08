package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"text/tabwriter"
	"time"

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
	"github.com/creativeprojects/resticprofile/util/bools"
	"github.com/creativeprojects/resticprofile/util/shutdown"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/spf13/pflag"
)

// These fields are populated by the goreleaser build
var (
	version = "0.23.0-dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func init() {
	rand.Seed(time.Now().UnixNano() - time.Now().Unix())
}

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
		fmt.Println(flagErr)
		_ = displayHelpCommand(os.Stdout, commandRequest{ownCommands: ownCommands, flags: flags, args: args})
		exitCode = 2
		return
	}

	if flags.wait {
		// keep the console running at the end of the program
		// so we can see what's going on
		defer func() {
			fmt.Println("\n\nPress the Enter Key to continue...")
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
	if flags.help || flagErr == pflag.ErrHelp {
		_ = displayHelpCommand(os.Stdout, commandRequest{ownCommands: ownCommands, flags: flags, args: args})
		return
	}

	// setting up the logger - we can start logging right after
	if flags.isChild {
		// use a remote logger
		client := remote.NewClient(flags.parentPort)
		setupRemoteLogger(flags, client)

		// also redirect the terminal through the client
		term.SetAllOutput(term.NewRemoteTerm(client))

		// If this is running in elevated mode we'll need to send a finished signal
		if flags.isChild {
			defer func(port int) {
				client := remote.NewClient(port)
				client.Done()
			}(flags.parentPort)
		}

	} else if flags.log != "" {
		handle, err := setupTargetLogger(flags)
		if err != nil {
			// back to a console logger
			setupConsoleLogger(flags)
			clog.Errorf("cannot open log target: %s", err)
		} else {
			// close the log file at the end
			defer handle.Close()
		}

	} else {
		// Use the console logger
		setupConsoleLogger(flags)
	}

	// keep this one last if possible (so it will be first at the end)
	defer showPanicData()

	banner()

	// resticprofile own commands (configuration file NOT loaded)
	if len(flags.resticArgs) > 0 {
		if ownCommands.Exists(flags.resticArgs[0], false) {
			err = ownCommands.Run(nil, flags.resticArgs[0], flags, flags.resticArgs[1:])
			if err != nil {
				clog.Error(err)
				exitCode = 1
				return
			}
			return
		}
	}

	configFile, err := filesearch.FindConfigurationFile(flags.config)
	if err != nil {
		clog.Error(err)
		exitCode = 1
		return
	}
	if configFile != flags.config {
		clog.Infof("using configuration file: %s", configFile)
	}

	c, err := config.LoadFile(configFile, flags.format)
	if err != nil {
		clog.Errorf("cannot load configuration file: %v", err)
		exitCode = 1
		return
	}

	global, err := c.GetGlobalSection()
	if err != nil {
		clog.Errorf("cannot load global configuration: %v", err)
		exitCode = 1
		return
	}

	// prevent computer from sleeping
	var caffeinate *preventsleep.Caffeinate
	if global.PreventSleep {
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
	if global.MinMemory > 0 {
		avail := free()
		if avail > 0 && avail < global.MinMemory {
			clog.Errorf("available memory is < %v MB (option 'min-memory' in the 'global' section)", global.MinMemory)
			exitCode = 1
			return
		}
	}

	if !flags.noPriority {
		err = setPriority(global.Nice, global.Priority)
		if err != nil {
			clog.Warning(err)
		}

		if global.IONice {
			err = priority.SetIONice(global.IONiceClass, global.IONiceLevel)
			if err != nil {
				clog.Warning(err)
			}
		}
	}

	resticBinary, err := filesearch.FindResticBinary(global.ResticBinary)
	if err != nil {
		clog.Error("cannot find restic: ", err)
		clog.Warning("you can specify the path of the restic binary in the global section of the configuration file (restic-binary)")
		exitCode = 1
		return
	}

	// The remaining arguments are going to be sent to the restic command line
	resticArguments := flags.resticArgs
	resticCommand := global.DefaultCommand
	if len(resticArguments) > 0 {
		resticCommand = resticArguments[0]
		resticArguments = resticArguments[1:]
	}

	// resticprofile own commands (with configuration file)
	if ownCommands.Exists(resticCommand, true) {
		err = ownCommands.Run(c, resticCommand, flags, resticArguments)
		if err != nil {
			clog.Error(err)
			exitCode = 1
			return
		}
		return
	}

	// detect restic version
	if len(global.ResticVersion) == 0 {
		if global.ResticVersion, err = restic.GetVersion(resticBinary); err != nil {
			clog.Warningf("assuming restic is at latest known version ; %s", err.Error())
			global.ResticVersion = restic.AnyVersion
		}
	}
	clog.Debugf("restic %s", global.ResticVersion)

	if c.HasProfile(flags.name) {
		// if running as a systemd timer
		notifyStart()
		defer notifyStop()

		// Single profile run
		err = runProfile(c, global, flags, flags.name, resticBinary, resticArguments, resticCommand, "")
		if err != nil {
			clog.Error(err)
			exitCode = 1
			return
		}

	} else if c.HasProfileGroup(flags.name) {
		// Group run
		group, err := c.GetProfileGroup(flags.name)
		if err != nil {
			clog.Errorf("cannot load group '%s': %v", flags.name, err)
		}
		if group != nil && len(group.Profiles) > 0 {
			// if running as a systemd timer
			notifyStart()
			defer notifyStop()

			for i, profileName := range group.Profiles {
				clog.Debugf("[%d/%d] starting profile '%s' from group '%s'", i+1, len(group.Profiles), profileName, flags.name)
				err = runProfile(c, global, flags, profileName, resticBinary, resticArguments, resticCommand, flags.name)
				if err != nil {
					clog.Error(err)
					if global.GroupContinueOnError && bools.IsTrueOrUndefined(group.ContinueOnError) ||
						bools.IsTrue(group.ContinueOnError) {
						// keep going to the next profile
						continue
					}
					exitCode = 1
					return
				}
			}
		}

	} else {
		clog.Errorf("profile or group not found '%s'", flags.name)
		displayProfiles(os.Stdout, c, flags)
		displayGroups(os.Stdout, c, flags)
		exitCode = 1
		return
	}
}

func banner() {
	clog.Debugf("resticprofile %s compiled with %s", version, runtime.Version())
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
		cleanup = func() {}
	}
	return
}

func runProfile(
	c *config.Config,
	global *config.Global,
	flags commandLineFlags,
	profileName string,
	resticBinary string,
	resticArguments []string,
	resticCommand string,
	group string,
) error {
	profile, cleanup, err := openProfile(c, profileName)
	defer cleanup()
	if err != nil {
		return err
	}

	displayProfileDeprecationNotices(profile)
	c.DisplayConfigurationIssues()

	// Send the quiet/verbose down to restic as well (override profile configuration)
	if flags.quiet {
		profile.Quiet = true
		profile.Verbose = constants.VerbosityNone
	}
	if flags.verbose {
		profile.Verbose = constants.VerbosityLevel1
		profile.Quiet = false
	}
	if flags.veryVerbose {
		profile.Verbose = constants.VerbosityLevel3
		profile.Quiet = false
	}

	// change log filter according to profile settings
	if profile.Quiet {
		changeLevelFilter(clog.LevelWarning)
	} else if profile.Verbose > constants.VerbosityNone && !flags.veryVerbose {
		changeLevelFilter(clog.LevelDebug)
	}

	// use the broken arguments escaping (before v0.15.0)
	if global.LegacyArguments {
		profile.SetLegacyArg(true)
	}

	// tell the profile what version of restic is in use
	if e := profile.SetResticVersion(global.ResticVersion); e != nil {
		clog.Warningf("restic version %q is no valid semver: %s", global.ResticVersion, e.Error())
	}

	// Specific case for the "host" flag where an empty value should be replaced by the hostname
	hostname := "none"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}
	profile.SetHost(hostname)

	// Catch CTR-C keypress, or other signal sent by a service manager (systemd)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	// remove signal catch before leaving
	defer signal.Stop(sigChan)

	wrapper := newResticWrapper(
		global,
		resticBinary,
		flags.dryRun,
		profile,
		resticCommand,
		resticArguments,
		sigChan,
	)

	if flags.noLock {
		wrapper.ignoreLock()
	} else if flags.lockWait > 0 {
		wrapper.maxWaitOnLock(flags.lockWait)
	}

	// add progress receivers if necessary
	if profile.StatusFile != "" {
		wrapper.addProgress(status.NewProgress(profile, status.NewStatus(profile.StatusFile)))
	}
	if profile.PrometheusPush != "" || profile.PrometheusSaveToFile != "" {
		wrapper.addProgress(prom.NewProgress(profile, prom.NewMetrics(group, version, profile.PrometheusLabels)))
	}

	err = wrapper.runProfile()
	if err != nil {
		return err
	}
	return nil
}

// randomBool returns true for Heads and false for Tails
func randomBool() bool {
	return rand.Int31n(10000) < 5000
}

func free() uint64 {
	mem, err := memory.Get()
	if err != nil {
		clog.Info("OS memory information not available")
		return 0
	}
	avail := (mem.Total - mem.Used) / 1048576
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
		if goarm > 0 {
			_, _ = fmt.Fprintf(w, "\t%s:\tv%d\n", "arm", goarm)
		}
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
