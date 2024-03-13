//go:build !windows && !linux
// +build !windows,!linux

package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/priority"
)

// This is only displaying the priority of the current process (for testing)
func main() {
	processNice, err := priority.GetProcessNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	groupNice, err := priority.GetGroupNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Process Priority: %d, Group Priority: %d\n", processNice, groupNice)
}
