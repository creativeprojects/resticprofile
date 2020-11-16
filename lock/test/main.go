package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/creativeprojects/resticprofile/lock"
)

// quick command to put a lock, wait for some time, then release the lock

func main() {
	wait := 0
	lockfile := ""
	flag.IntVar(&wait, "wait", 1000, "Wait n milliseconds before unlocking")
	flag.StringVar(&lockfile, "lock", "test.lock", "Name of the lock file")
	flag.Parse()

	l := lock.NewLock(lockfile)
	if l.TryAcquire() {
		defer func() {
			l.Release()
			fmt.Println("lock released")
		}()

		// Catch CTR-C key
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
		// remove signal catch before leaving
		defer signal.Stop(sigChan)

		fmt.Println("lock acquired")

		select {
		case <-sigChan:
			fmt.Println("task interrupted")
		case <-time.After(time.Duration(wait) * time.Millisecond):
			fmt.Println("task finished")
		}

	} else {
		fmt.Println("cannot acquire lock")
		os.Exit(1)
	}
}
