//go:build windows

package main

import (
	"fmt"

	"github.com/creativeprojects/resticprofile/priority"
	"golang.org/x/sys/windows"
)

// This is only displaying the priority of the current process (for testing)
func runPriority() int {
	handle := windows.CurrentProcess()
	class, err := windows.GetPriorityClass(handle)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf("Priority class: %s\n", priority.GetPriorityClassName(class))
	return 0
}
