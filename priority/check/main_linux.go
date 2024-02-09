//go:build linux

package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/priority"
)

// This is only displaying the priority of the current process (for testing)
func main() {
	getPriority()
	getIOPriority()
}

func getPriority() {
	pri, err := priority.GetNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Priority: %d\n", pri)
}

func getIOPriority() {
	class, value, err := priority.GetIONice()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("IOPriority: class = %d, value = %d\n", class, value)
}
