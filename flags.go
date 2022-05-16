package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/pflag"
)

type commandLineFlags struct {
	help        bool
	quiet       bool
	verbose     bool
	veryVerbose bool
	config      string
	format      string
	name        string
	logFile     string
	dryRun      bool
	noLock      bool
	lockWait    time.Duration
	noAnsi      bool
	theme       string
	resticArgs  []string
	selfUpdate  bool
	wait        bool
	isChild     bool
	parentPort  int
	noPriority  bool
	run         string
}

// loadFlags loads command line flags (before any command)
func loadFlags(args []string) (*pflag.FlagSet, commandLineFlags, error) {
	flagset := pflag.NewFlagSet("resticprofile", pflag.ContinueOnError)

	flagset.Usage = func() {
		fmt.Println("\nUsage of resticprofile:")
		fmt.Println("\tresticprofile [resticprofile flags] [profile name.][restic command] [restic flags]")
		fmt.Println("\tresticprofile [resticprofile flags] [profile name.][resticprofile command] [command specific flags]")
		fmt.Println("\nresticprofile flags:")
		flagset.PrintDefaults()
		fmt.Println("\nresticprofile own commands:")
		displayOwnCommands(os.Stdout)
		fmt.Println("\nDocumentation available at https://creativeprojects.github.io/resticprofile/")
		fmt.Println("")
	}

	flags := commandLineFlags{}

	flagset.BoolVarP(&flags.help, "help", "h", false, "display this help")
	flagset.BoolVarP(&flags.quiet, "quiet", "q", constants.DefaultQuietFlag, "display only warnings and errors")
	flagset.BoolVarP(&flags.verbose, "verbose", "v", constants.DefaultVerboseFlag, "display some debugging information")
	flagset.BoolVar(&flags.veryVerbose, "trace", constants.DefaultVerboseFlag, "display even more debugging information")
	flagset.StringVarP(&flags.config, "config", "c", constants.DefaultConfigurationFile, "configuration file")
	flagset.StringVarP(&flags.format, "format", "f", "", "file format of the configuration (default is to use the file extension)")
	flagset.StringVarP(&flags.name, "name", "n", constants.DefaultProfileName, "profile name")
	flagset.StringVarP(&flags.logFile, "log", "l", "", "logs into a file instead of the console")
	flagset.BoolVar(&flags.dryRun, "dry-run", false, "display the restic commands instead of running them")

	flagset.BoolVar(&flags.noLock, "no-lock", false, "skip profile lock file")
	flagset.DurationVar(&flags.lockWait, "lock-wait", 0, "wait up to duration to acquire a lock (syntax \"1h5m30s\")")

	flagset.BoolVar(&flags.noAnsi, "no-ansi", false, "disable ansi control characters (disable console colouring)")
	flagset.StringVar(&flags.theme, "theme", constants.DefaultTheme, "console colouring theme (dark, light, none)")
	flagset.BoolVar(&flags.noPriority, "no-prio", false, "don't set any priority on load: used when started from a service that has already set the priority")

	flagset.BoolVarP(&flags.wait, "wait", "w", false, "wait at the end until the user presses the enter key")

	if runtime.GOOS == "windows" {
		// flag for internal use only
		flagset.BoolVar(&flags.isChild, constants.FlagAsChild, false, "run as an elevated user child process")
		flagset.IntVar(&flags.parentPort, constants.FlagPort, 0, "port of the parent process")
		_ = flagset.MarkHidden(constants.FlagAsChild)
		_ = flagset.MarkHidden(constants.FlagPort)
	}

	// Deprecated since 0.7.0
	flagset.BoolVar(&flags.selfUpdate, "self-update", false, "auto update of resticprofile (does not update restic)")
	_ = flagset.MarkHidden("self-update")

	// stop at the first non flag found; the rest will be sent to the restic command line
	flagset.SetInterspersed(false)

	err := flagset.Parse(args)
	if err != nil {
		return flagset, flags, err
	}

	// remaining flags
	flags.resticArgs = flagset.Args()

	// if there are no further arguments, no further parsing is needed
	if len(flags.resticArgs) == 0 {
		return flagset, flags, err
	}

	// parse first postitional argument as <profile>.<command> if the profile was not set via name
	nameFlag := flagset.Lookup("name")
	if (nameFlag == nil || !nameFlag.Changed) && strings.Contains(flags.resticArgs[0], ".") {
		// split first argument at `.`
		profileAndCommand := strings.Split(flags.resticArgs[0], ".")
		// last element will be used as restic command
		command := profileAndCommand[len(profileAndCommand)-1]
		// remaining elements will be stiched together with `.` and used as profile name
		profile := strings.Join(profileAndCommand[0:len(profileAndCommand)-1], ".")

		// set command
		if len(command) == 0 {
			// take default command by removing it from resticArgs
			flags.resticArgs = flags.resticArgs[1:]
		} else {
			flags.resticArgs[0] = command
		}

		// set profile
		if len(profile) == 0 {
			profile = constants.DefaultProfileName
		}
		flags.name = profile
	}

	return flagset, flags, nil
}
