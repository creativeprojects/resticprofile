package main

import (
	"fmt"
	"os"
	"runtime"

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
	noAnsi      bool
	theme       string
	resticArgs  []string
	selfUpdate  bool
	wait        bool
	isChild     bool
	parentPort  int
}

// loadFlags loads command line flags (before any command)
func loadFlags() (*pflag.FlagSet, commandLineFlags) {
	flagset := pflag.NewFlagSet("resticprofile", pflag.ExitOnError)

	flagset.Usage = func() {
		fmt.Println("\nUsage of resticprofile:")
		fmt.Println("\tresticprofile [resticprofile flags] [command] [restic flags]")
		fmt.Println("\nresticprofile flags:")
		flagset.PrintDefaults()
		fmt.Println("\nresticprofile own commands:")
		displayOwnCommands()
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

	flagset.BoolVar(&flags.noAnsi, "no-ansi", false, "disable ansi control characters (disable console colouring)")
	flagset.StringVar(&flags.theme, "theme", constants.DefaultTheme, "console colouring theme (dark, light, none)")

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

	_ = flagset.Parse(os.Args[1:])

	// remaining flags
	flags.resticArgs = flagset.Args()

	return flagset, flags
}
