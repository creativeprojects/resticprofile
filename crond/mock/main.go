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

	if list {
		if noCrontab != "" {
			if noCrontab == "user" {
				println("no crontab for user")
				os.Exit(1)
			} else {
				return
			}
		} else if crontab != "" {
			print(crontab)
		} else {
			os.Exit(6)
		}
	} else {
		if slices.Contains(os.Args, "-") {
			if crontab != "" {
				sb := new(strings.Builder)
				if _, err := io.Copy(sb, os.Stdin); err == nil {
					if stdin := strings.TrimSpace(sb.String()); stdin != crontab {
						err = fmt.Errorf("%q != %q", crontab, stdin)
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
							// os.Exit(1)
						}
					}
				}
			} else {
				os.Exit(8)
			}
		} else {
			os.Exit(10)
		}
	}
}
