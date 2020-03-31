//+build windows

package main

import (
	"fmt"
	"os"

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
	switch class {
	case windows.ABOVE_NORMAL_PRIORITY_CLASS:
		fmt.Printf("Priority class: ABOVE_NORMAL")
	case windows.BELOW_NORMAL_PRIORITY_CLASS:
		fmt.Printf("Priority class: BELOW_NORMAL")
	case windows.HIGH_PRIORITY_CLASS:
		fmt.Printf("Priority class: HIGH")
	case windows.IDLE_PRIORITY_CLASS:
		fmt.Printf("Priority class: IDLE")
	case windows.NORMAL_PRIORITY_CLASS:
		fmt.Printf("Priority class: NORMAL")
	case windows.PROCESS_MODE_BACKGROUND_BEGIN:
		fmt.Printf("Priority class: PROCESS_MODE_BACKGROUND (begin)")
	case windows.PROCESS_MODE_BACKGROUND_END:
		fmt.Printf("Priority class: PROCESS_MODE_BACKGROUND (end)")
	case windows.REALTIME_PRIORITY_CLASS:
		fmt.Printf("Priority class: REALTIME")
	default:
		fmt.Printf("Priority class: 0x%x\n", class)
	}
}
