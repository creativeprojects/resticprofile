package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creativeprojects/resticprofile/lock"
)

func runLock(wait int, lockfile string) int {
	sigChan := make(chan os.Signal, 2)
	signal.Ignore()
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	// remove signal catch before leaving
	defer signal.Stop(sigChan)

	fmt.Println("started")

	l := lock.NewLock(lockfile)
	if l.TryAcquire() {
		defer func() {
			l.Release()
			fmt.Println("lock released")
		}()

		fmt.Println("lock acquired")

		select {
		case <-sigChan:
			fmt.Println("task interrupted")
		case <-time.After(time.Duration(wait) * time.Millisecond):
			fmt.Println("task finished")
		}
		return 0
	}

	fmt.Println("cannot acquire lock")
	return 1
}
