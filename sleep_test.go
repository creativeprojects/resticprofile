package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNonInterruptedSleep(t *testing.T) {
	err := interruptibleSleep(30*time.Millisecond, make(<-chan os.Signal))
	assert.NoError(t, err)
}

func TestInterruptedSleep(t *testing.T) {
	maxWait := 3 * time.Second
	sigChan := make(chan os.Signal)
	go func() {
		time.Sleep(30 * time.Millisecond)
		sigChan <- os.Interrupt
	}()
	start := time.Now()
	err := interruptibleSleep(maxWait, sigChan)
	assert.ErrorIs(t, err, errInterrupt)
	assert.Less(t, time.Since(start), maxWait)
}
