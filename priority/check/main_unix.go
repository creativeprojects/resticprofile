//go:build !windows && !linux

package main

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

const selfPID = 0

// This is only displaying the priority of the current process (for testing)
func main() {
	// run it in a go routine in case it would make a difference
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		displayProcessAndGroup()
		displayPriority()
	}()
	wg.Wait()
}

func displayProcessAndGroup() {
	fmt.Printf("Process ID: %d, Group ID: %d\n", unix.Getpid(), unix.Getpgrp())
}

func displayPriority() {
	processNice, err := getProcessNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	groupNice, err := getGroupNice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Process Priority: %d, Group Priority: %d\n", processNice, groupNice)
}

// getProcessNice returns the nice value of the current process
func getProcessNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PROCESS, selfPID)
	if err != nil {
		return 0, err
	}
	return pri, nil
}

// GetProcessNice returns the nice value of the current process group
func getGroupNice() (int, error) {
	pri, err := unix.Getpriority(unix.PRIO_PGRP, selfPID)
	if err != nil {
		return 0, err
	}
	return pri, nil
}
