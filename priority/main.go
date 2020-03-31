//+build !windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main() {
	pid := unix.Getpid()
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, pid)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("PID: %d, Priority: %d\n", pid, pri)
}
