package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/creativeprojects/resticprofile/term"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/priority"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/mackerelio/go-osstat/memory"
)

// These fields are populated by the goreleaser build
var (
	version = "0.10.0-dev"
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

	flagset, flags := loadFlags()

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
	if flags.help {
		flagset.Usage()
		return
	}

	// setting up the logger - we can start sending messages right after
	if flags.isChild {
		// use a remote logger
		client := remote.NewClient(flags.parentPort)
		setupRemoteLogger(client)

		// also redirect the terminal through the client
		term.SetAllOutput(term.NewRemoteTerm(client))

		// If this is running in elevated mode we'll need to send a finished signal
		if flags.isChild {
			defer func(port int) {
				client := remote.NewClient(port)
				client.Done()
			}(flags.parentPort)
		}

	} else if flags.logFile != "" {
		file, err := setupFileLogger(flags)
		if err != nil {
			// back to a console logger
			setupConsoleLogger(flags)
			clog.Errorf("cannot open logfile: %s", err)
		} else {
			// also redirect all terminal output
			term.SetAllOutput(file)
			// close the log file at the end
			defer file.Close()
		}

	} else {
		// Use the console logger
		setupConsoleLogger(flags)
	}

	// keep this one last if possible (so it will be first at the end)
	defer showPanicData()

	banner()

	// Deprecated in version 0.7.0
	// Keep for compatibility with version 0.6.1
	if flags.selfUpdate {
		err = confirmAndSelfUpdate(flags.verbose)
		if err != nil {
			clog.Error(err)
			exitCode = 1
			return
		}
		return
	}

	// resticprofile own commands (configuration file NOT loaded)
	if len(flags.resticArgs) > 0 {
		if isOwnCommand(flags.resticArgs[0], false) {
			err = runOwnCommand(nil, flags.resticArgs[0], flags, flags.resticArgs[1:])
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

	// Check memory pressure
	if global.MinMemory > 0 {
		avail := free()
		if avail > 0 && avail < global.MinMemory {
			clog.Errorf("available memory is < %v MB (option 'min-memory' in the 'global' section)", global.MinMemory)
			exitCode = 1
			return
		}
	}

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
	if isOwnCommand(resticCommand, true) {
		err = runOwnCommand(c, resticCommand, flags, resticArguments)
		if err != nil {
			clog.Error(err)
			exitCode = 1
			return
		}
		return
	}

	if c.HasProfile(flags.name) {
		// Single profile run
		err = runProfile(c, global, flags, flags.name, resticBinary, resticArguments, resticCommand)
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
		if len(group) > 0 {
			for i, profileName := range group {
				clog.Debugf("[%d/%d] starting profile '%s' from group '%s'", i+1, len(group), profileName, flags.name)
				err = runProfile(c, global, flags, profileName, resticBinary, resticArguments, resticCommand)
				if err != nil {
					clog.Error(err)
					exitCode = 1
					return
				}
			}
		}

	} else {
		clog.Errorf("profile or group not found '%s'", flags.name)
		displayProfiles(c)
		displayGroups(c)
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

func runProfile(
	c *config.Config,
	global *config.Global,
	flags commandLineFlags,
	profileName string,
	resticBinary string,
	resticArguments []string,
	resticCommand string,
) error {
	var err error

	profile, err := c.GetProfile(profileName)
	if err != nil {
		clog.Warning(err)
	}
	if profile == nil {
		return fmt.Errorf("cannot load profile '%s'", profileName)
	}

	// Send the quiet/verbose down to restic as well (override profile configuration)
	if flags.quiet {
		profile.Quiet = true
		profile.Verbose = false
	}
	if flags.verbose {
		profile.Verbose = true
		profile.Quiet = false
	}

	// All files in the configuration are relative to the configuration file, NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	if rootPath != "." {
		clog.Debugf("files in configuration are relative to '%s'", rootPath)
	}
	profile.SetRootPath(rootPath)

	// Specific case for the "host" flag where an empty value should be replaced by the hostname
	hostname := "none"
	currentHost, err := os.Hostname()
	if err == nil {
		hostname = currentHost
	}
	profile.SetHost(hostname)

	// Catch CTR-C keypress
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)

	wrapper := newResticWrapper(
		resticBinary,
		global.Initialize || profile.Initialize,
		flags.dryRun,
		profile,
		resticCommand,
		resticArguments,
		sigChan,
	)
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
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "version", version)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "commit", commit)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "compiled", date)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "by", builtBy)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "error", r)
		_, _ = fmt.Fprintf(w, "\t%s:\t%s\n", "stack", string(debug.Stack()))
		w.Flush()
		fmt.Fprint(os.Stderr, "===============================================================\n")
	}
}
