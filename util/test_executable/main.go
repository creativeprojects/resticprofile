package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/util"
)

func main() {
	executable, err := util.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("%q\n", executable)
}
