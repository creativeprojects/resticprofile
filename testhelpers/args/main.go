package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/util"
)

func main() {
	exitCode := run()
	if exitCode > 0 {
		os.Exit(exitCode)
	}
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Println("missing command argument")
		return 1
	}

	switch os.Args[1] {

	case "executable":
		executable, err := util.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting executable: %s\n", err)
			return 1
		}
		fmt.Printf("%q\n", executable)
		return 0

	case "lock":
		wait := 0
		lockfile := ""
		flags := flag.NewFlagSet("lock", flag.ContinueOnError)
		flags.IntVar(&wait, "wait", 1000, "Wait n milliseconds before unlocking")
		flags.StringVar(&lockfile, "lock", "test.lock", "Name of the lock file")
		if err := flags.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "parsing flags: %s", err)
			return 1
		}

		return runLock(wait, lockfile)

	case "priority":
		return runPriority()

	default:
		fmt.Printf("command argument %q not supported\n", os.Args[1])
		return 1
	}
}
