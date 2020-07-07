package clog

import (
	"bytes"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	SetTestLog(t)
	defer ClearTestLog()

	Debug("one", "two", "three")
	Info("one", "two", "three")
	Warning("one", "two", "three")
	Error("one", "two", "three")

	Debugf("%d %d %d", 1, 2, 3)
	Infof("%d %d %d", 1, 2, 3)
	Warningf("%d %d %d", 1, 2, 3)
	Errorf("%d %d %d", 1, 2, 3)
}

func TestFileLoggerConcurrency(t *testing.T) {
	// remove date prefix on logs
	log.SetFlags(0)

	iterations := 1000
	buffer := &bytes.Buffer{}
	logger := NewStreamLog(buffer)
	wg := sync.WaitGroup{}
	wg.Add(iterations)
	for i := 0; i < iterations; i++ {
		go func(num int) {
			logger.Infof("log %03d", num)
			wg.Done()
		}(i)
	}
	wg.Wait()
	for line, err := buffer.ReadString('\n'); err == nil; line, err = buffer.ReadString('\n') {
		assert.Len(t, line, 14)
	}
	// clean up
	log.SetFlags(log.LstdFlags)
}
