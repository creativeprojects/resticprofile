package main

import (
	"flag"
	"fmt"
)

type commandLineFlags struct {
	help    bool
	quiet   bool
	verbose bool
	config  string
	name    string
	noAnsi  bool
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

	flag.BoolVar(&flags.quiet, "q", false, "display only warnings and errors - shorthand")
	flag.BoolVar(&flags.quiet, "quiet", false, "display only warnings and errors")

	flag.BoolVar(&flags.verbose, "v", false, "display debugging information - shorthand")
	flag.BoolVar(&flags.verbose, "verbose", false, "display debugging information")

	flag.StringVar(&flags.config, "c", "profiles.conf", "configuration file - shorthand")
	flag.StringVar(&flags.config, "config", "profiles.conf", "configuration file")

	flag.StringVar(&flags.name, "n", "default", "profile name - shorthand")
	flag.StringVar(&flags.name, "name", "default", "profile name")

	flag.BoolVar(&flags.noAnsi, "no-ansi", false, "disable console colouring")

	flag.Parse()
	return flags
}
