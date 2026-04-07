package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
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

	code := run(wait, lockfile)
	if code > 0 {
		os.Exit(code)
	}
}

func run(wait int, lockfile string) int {
	// Catch user defined signals
	sigChan := make(chan os.Signal, 2)
	signal.Ignore()
	signal.Notify(sigChan, os.Interrupt)
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
