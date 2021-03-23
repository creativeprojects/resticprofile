package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

// Implementation of a mock command that mimics the output of restic during tests
func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "not enough arguments\n")
		os.Exit(2)
	}

	// first argument is a command
	command := os.Args[1]

	stderr := ""
	exit := 0
	arguments := false
	sleep := 0
	flags := flag.NewFlagSet("mock", flag.ExitOnError)
	flags.StringVar(&stderr, "stderr", "", "send this message to stderr")
	flags.IntVar(&exit, "exit", 0, "set exit code")
	flags.BoolVar(&arguments, "args", false, "display command line arguments")
	flags.IntVar(&sleep, "sleep", 0, "sleep timer in ms")
	flags.Parse(os.Args[2:])

	if arguments {
		// echo command and arguments to stdout
		fmt.Printf("args: %q\n", os.Args[1:])
		fmt.Printf("command: %s\n", command)
	}

	if stderr != "" {
		fmt.Fprintf(os.Stderr, "%s\n", stderr)
	}

	if sleep > 0 {
		time.Sleep(time.Duration(sleep * int(time.Millisecond)))
	}
	os.Exit(exit)
}
