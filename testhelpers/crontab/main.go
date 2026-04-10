package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

func main() {
	var list bool
	flag.BoolVar(&list, "l", false, "list")
	flag.Parse()
	noCrontab := os.Getenv("NO_CRONTAB")
	crontab := os.Getenv("CRONTAB")

	exitCode := run(list, noCrontab, crontab)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func run(list bool, noCrontab, crontab string) int {
	if list {
		if noCrontab != "" {
			if noCrontab == "user" {
				println("no crontab for user")
				return 1
			} else {
				return 0
			}
		} else if crontab != "" {
			print(crontab)
		} else {
			return 6
		}
		return 0
	}

	if slices.Contains(os.Args, "-") {
		if crontab != "" {
			sb := new(strings.Builder)
			if _, err := io.Copy(sb, os.Stdin); err == nil {
				if stdin := strings.TrimSpace(sb.String()); stdin != crontab {
					fmt.Fprintf(os.Stderr, "%q != %q\n", crontab, stdin)
				}
			}
		} else {
			return 8
		}
		return 0
	}

	return 10
}
