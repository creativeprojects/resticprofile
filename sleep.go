package main

import (
	"errors"
	"os"
	"time"
)

var (
	errInterrupt = errors.New("interrupted")
)

// interruptibleSleep returns errInterrupt if interrupted by the channel.
func interruptibleSleep(delay time.Duration, interrupt <-chan os.Signal) error {
	select {
	case <-time.After(delay):
		return nil
	case <-interrupt:
		return errInterrupt
	}
}
