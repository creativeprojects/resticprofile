package main

import (
	"flag"
	"fmt"

	"github.com/creativeprojects/resticprofile/constants"
)

type commandLineFlags struct {
	help    bool
	quiet   bool
	verbose bool
	config  string
	name    string
	noAnsi  bool
	theme   string
}

func loadFlags() commandLineFlags {
	flag.Usage = func() {
		fmt.Println("\nUsage of resticprofile:")
		fmt.Println("\tresticprofile [resticprofile flags] [command] [restic flags]")
		fmt.Println("\nresticprofile flags:")
		flag.PrintDefaults()
		fmt.Println("")
	}

	flags := commandLineFlags{}

	flag.BoolVar(&flags.help, "h", false, "display this help - shorthand")
	flag.BoolVar(&flags.help, "help", false, "display this help")

	flag.BoolVar(&flags.quiet, "q", constants.DefaultQuietFlag, "display only warnings and errors - shorthand")
	flag.BoolVar(&flags.quiet, "quiet", constants.DefaultQuietFlag, "display only warnings and errors")

	flag.BoolVar(&flags.verbose, "v", constants.DefaultVerboseFlag, "display debugging information - shorthand")
	flag.BoolVar(&flags.verbose, "verbose", constants.DefaultVerboseFlag, "display debugging information")

	flag.StringVar(&flags.config, "c", constants.DefaultConfigurationFile, "configuration file - shorthand")
	flag.StringVar(&flags.config, "config", constants.DefaultConfigurationFile, "configuration file")

	flag.StringVar(&flags.name, "n", constants.DefaultProfileName, "profile name - shorthand")
	flag.StringVar(&flags.name, "name", constants.DefaultProfileName, "profile name")

	flag.BoolVar(&flags.noAnsi, "no-ansi", false, "disable ansi control characters (used for console colouring)")
	flag.StringVar(&flags.theme, "theme", constants.DefaultTheme, "colouring theme (dark, light, none)")

	flag.Parse()
	return flags
}
