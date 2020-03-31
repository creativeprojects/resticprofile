//+build windows

package main

import (
	"fmt"
	"os"

	"github.com/creativeprojects/resticprofile/priority"
	"golang.org/x/sys/windows"
)

// This is only displaying the priority of the current process (for testing)
func main() {
	handle, err := windows.GetCurrentProcess()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	class, err := windows.GetPriorityClass(handle)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Priority class: %s\n", priority.GetPriorityClassName(class))
}
