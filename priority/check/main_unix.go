//+build !windows,!linux

package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/priority"
)

// This is only displaying the priority of the current process (for testing)
func main() {
	pri, err := priority.GetNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Priority: %d\n", pri)
}
