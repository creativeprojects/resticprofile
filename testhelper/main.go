package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(`¯\_(ツ)_/¯`)
		return
	}

	switch os.Args[1] {
	case "executable":
		executable, err := util.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting executable: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%q\n", executable)
		return

	default:
		fmt.Println(`¯\_(ツ)_/¯`)
		return
	}
}
