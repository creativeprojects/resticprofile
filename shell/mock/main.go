package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
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
	flags := flag.NewFlagSet("mock", flag.ContinueOnError)
	flags.StringVar(&stderr, "stderr", "", "send this message to stderr")
	flags.IntVar(&exit, "exit", 0, "set exit code")
	flags.BoolVar(&arguments, "args", false, "display command line arguments")
	flags.IntVar(&sleep, "sleep", 0, "sleep timer in ms")

	if err := flags.Parse(os.Args[2:]); err != nil {
		if err == flag.ErrHelp {
			return
		} else if !strings.Contains(err.Error(), "flag provided but not defined") {
			os.Exit(1)
		}
	}

	if arguments {
		// echo command and arguments to stdout
		fmt.Printf("args: %q\n", os.Args[1:])
		fmt.Printf("command: %s\n", command)
	}

	if stderr != "" {
		if strings.HasPrefix(stderr, "@") {
			if file, err := os.Open(stderr[1:]); err != nil {
				stderr = err.Error()
				exit = 3
			} else {
				io.CopyN(os.Stderr, file, 1024)
				file.Close()
				stderr = ""
			}
		}

		fmt.Fprintf(os.Stderr, "%s\n", stderr)
	}

	if sleep > 0 {
		time.Sleep(time.Duration(sleep * int(time.Millisecond)))
	}
	os.Exit(exit)
}
