package main

import (
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

	default:
		fmt.Printf("command argument %q not supported\n", os.Args[1])
		return 1
	}
}
