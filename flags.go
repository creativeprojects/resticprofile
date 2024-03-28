package main

import (
	"os"
	"slices"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
)

type commandLineFlags struct {
	help            bool
	quiet           bool
	verbose         bool
	veryVerbose     bool
	config          string
	format          string
	name            string
	log             string // file path or log url
	commandOutput   string
	dryRun          bool
	noLock          bool
	lockWait        time.Duration
	noAnsi          bool
	theme           string
	resticArgs      []string
	wait            bool
	isChild         bool
	stderr          bool
	parentPort      int
	noPriority      bool
	ignoreOnBattery int
	usagesHelp      string
}

func envValueOverride[T any](defaultValue T, keys ...string) T {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); len(value) > 0 {
			var (
				err error
				v   any = defaultValue
			)
			switch v.(type) {
			case bool:
				v, err = cast.ToBoolE(value)
			case int:
				v, err = cast.ToIntE(value)
			case time.Duration:
				v, err = cast.ToDurationE(value)
			case string:
				v = value
			}
			if err == nil {
				defaultValue = v.(T)
			} else {
				clog.Errorf("cannot convert env variable %s=%q: %s", key, value, err.Error())
			}
			break
		}
	}
	return defaultValue
}

// loadFlags loads command line flags (before any command)
func loadFlags(args []string) (*pflag.FlagSet, commandLineFlags, error) {
	flagset := pflag.NewFlagSet("resticprofile", pflag.ContinueOnError)

	// flags with default values and env overrides
	flags := commandLineFlags{
		help:            false,
		quiet:           envValueOverride(constants.DefaultQuietFlag, "RESTICPROFILE_QUIET"),
		verbose:         envValueOverride(constants.DefaultVerboseFlag, "RESTICPROFILE_VERBOSE"),
		veryVerbose:     envValueOverride(constants.DefaultVerboseFlag, "RESTICPROFILE_TRACE"),
		config:          envValueOverride(constants.DefaultConfigurationFile, "RESTICPROFILE_CONFIG"),
		format:          envValueOverride("", "RESTICPROFILE_FORMAT"),
		name:            envValueOverride(constants.DefaultProfileName, "RESTICPROFILE_NAME"),
		log:             envValueOverride("", "RESTICPROFILE_LOG"),
		commandOutput:   envValueOverride(constants.DefaultCommandOutput, "RESTICPROFILE_COMMAND_OUTPUT"),
		dryRun:          envValueOverride(false, "RESTICPROFILE_DRY_RUN"),
		noLock:          envValueOverride(false, "RESTICPROFILE_NO_LOCK"),
		lockWait:        envValueOverride(time.Duration(0), "RESTICPROFILE_LOCK_WAIT"),
		stderr:          envValueOverride(false, "RESTICPROFILE_STDERR"),
		noAnsi:          envValueOverride(false, "RESTICPROFILE_NO_ANSI"),
		theme:           envValueOverride(constants.DefaultTheme, "RESTICPROFILE_THEME"),
		noPriority:      envValueOverride(false, "RESTICPROFILE_NO_PRIORITY"),
		wait:            envValueOverride(false, "RESTICPROFILE_WAIT"),
		ignoreOnBattery: envValueOverride(0, "RESTICPROFILE_IGNORE_ON_BATTERY"),
	}

	flagset.BoolVarP(&flags.help, "help", "h", flags.help, "display this help")
	flagset.BoolVarP(&flags.quiet, "quiet", "q", flags.quiet, "display only warnings and errors")
	flagset.BoolVarP(&flags.verbose, "verbose", "v", flags.verbose, "display some debugging information")
	flagset.BoolVar(&flags.veryVerbose, "trace", flags.veryVerbose, "display even more debugging information")
	flagset.StringVarP(&flags.config, "config", "c", flags.config, "configuration file")
	flagset.StringVarP(&flags.format, "format", "f", flags.format, "file format of the configuration (default is to use the file extension)")
	flagset.StringVarP(&flags.name, "name", "n", flags.name, "profile name")
	flagset.StringVarP(&flags.log, "log", "l", flags.log, "logs to a target instead of the console (file, syslog:[//server])")
	flagset.StringVar(&flags.commandOutput, "command-output", flags.commandOutput, "redirect command output when a log target is specified (log, console, all)")
	flagset.BoolVar(&flags.dryRun, "dry-run", flags.dryRun, "display the restic commands instead of running them")
	flagset.BoolVar(&flags.noLock, "no-lock", flags.noLock, "skip profile lock file")
	flagset.DurationVar(&flags.lockWait, "lock-wait", flags.lockWait, "wait up to duration to acquire a lock (syntax \"1h5m30s\")")
	flagset.BoolVar(&flags.stderr, "stderr", flags.noAnsi, "send console output to stderr (enabled for \"cat\" and \"dump\")")
	flagset.BoolVar(&flags.noAnsi, "no-ansi", flags.noAnsi, "disable ansi control characters (disable console colouring)")
	flagset.StringVar(&flags.theme, "theme", flags.theme, "console colouring theme (dark, light, none)")
	flagset.BoolVar(&flags.noPriority, "no-prio", flags.noPriority, "don't change the process priority: used when started from a service that has already set the priority")
	flagset.BoolVarP(&flags.wait, "wait", "w", flags.wait, "wait at the end until the user presses the enter key")
	flagset.IntVar(&flags.ignoreOnBattery, "ignore-on-battery", flags.ignoreOnBattery, "don't start the profile when the computer is running on battery. You can specify a value to ignore only when the % charge left is less or equal than the value")
	flagset.Lookup("ignore-on-battery").NoOptDefVal = "100" // 0 is flag not set, 100 is for a flag with no value (meaning just battery discharge)

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

	if err := flagset.Parse(args); err != nil {
		return flagset, flags, err
	}

	// remaining flags
	flags.resticArgs = flagset.Args()

	// if there are no further arguments, no further parsing is needed
	if len(flags.resticArgs) == 0 {
		return flagset, flags, nil
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
