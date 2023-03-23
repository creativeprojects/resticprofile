package main

import (
	"strings"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
)

type commandLineFlags struct {
	help        bool
	quiet       bool
	verbose     bool
	veryVerbose bool
	config      string
	format      string
	name        string
	log         string // file path or log url
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
	usagesHelp  string
}

// loadFlags loads command line flags (before any command)
func loadFlags(args []string) (*pflag.FlagSet, commandLineFlags, error) {
	flagset := pflag.NewFlagSet("resticprofile", pflag.ContinueOnError)

	flags := commandLineFlags{}

	flagset.BoolVarP(&flags.help, "help", "h", false, "display this help")
	flagset.BoolVarP(&flags.quiet, "quiet", "q", constants.DefaultQuietFlag, "display only warnings and errors")
	flagset.BoolVarP(&flags.verbose, "verbose", "v", constants.DefaultVerboseFlag, "display some debugging information")
	flagset.BoolVar(&flags.veryVerbose, "trace", constants.DefaultVerboseFlag, "display even more debugging information")
	flagset.StringVarP(&flags.config, "config", "c", constants.DefaultConfigurationFile, "configuration file")
	flagset.StringVarP(&flags.format, "format", "f", "", "file format of the configuration (default is to use the file extension)")
	flagset.StringVarP(&flags.name, "name", "n", constants.DefaultProfileName, "profile name")
	flagset.StringVarP(&flags.log, "log", "l", "", "logs to a target instead of the console")

	flagset.BoolVar(&flags.dryRun, "dry-run", false, "display the restic commands instead of running them")

	flagset.BoolVar(&flags.noLock, "no-lock", false, "skip profile lock file")
	flagset.DurationVar(&flags.lockWait, "lock-wait", 0, "wait up to duration to acquire a lock (syntax \"1h5m30s\")")

	flagset.BoolVar(&flags.noAnsi, "no-ansi", false, "disable ansi control characters (disable console colouring)")
	flagset.StringVar(&flags.theme, "theme", constants.DefaultTheme, "console colouring theme (dark, light, none)")
	flagset.BoolVar(&flags.noPriority, "no-prio", false, "don't set any priority on load: used when started from a service that has already set the priority")

	flagset.BoolVarP(&flags.wait, "wait", "w", false, "wait at the end until the user presses the enter key")

	if platform.IsWindows() {
		// flag for internal use only
		flagset.BoolVar(&flags.isChild, constants.FlagAsChild, false, "run as an elevated user child process")
		flagset.IntVar(&flags.parentPort, constants.FlagPort, 0, "port of the parent process")
		_ = flagset.MarkHidden(constants.FlagAsChild)
		_ = flagset.MarkHidden(constants.FlagPort)
	}

	// stop at the first non flag found; the rest will be sent to the restic command line
	flagset.SetInterspersed(false)

	// Store usage help for help command
	width, _ := term.OsStdoutTerminalSize()
	flags.usagesHelp = flagset.FlagUsagesWrapped(width)

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

	// handle explicit help request in command args (works with restic and own commands)
	if slices.ContainsFunc(flags.resticArgs, collect.In("--help", "-h")) {
		flags.help = true
	}

	// parse first positional argument as <profile>.<command> if the profile was not set via name
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
