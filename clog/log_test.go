package clog

import (
	"bytes"
	"log"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	SetTestLog(t)
	defer ClearTestLog()

	Log(NoLevel, "one", "two", "three")
	Debug("one", "two", "three")
	Info("one", "two", "three")
	Warning("one", "two", "three")
	Error("one", "two", "three")

	Logf(NoLevel, "%d %d %d", 1, 2, 3)
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

func TestLoggerVerbosity(t *testing.T) {
	expected := []string{
		"      0 >= 0",
		"DEBUG 1 >= 0",
		"INFO  2 >= 0",
		"WARN  3 >= 0",
		"ERROR 4 >= 0",
		"DEBUG 1 >= 1",
		"INFO  2 >= 1",
		"WARN  3 >= 1",
		"ERROR 4 >= 1",
		"INFO  2 >= 2",
		"WARN  3 >= 2",
		"ERROR 4 >= 2",
		"WARN  3 >= 3",
		"ERROR 4 >= 3",
		"ERROR 4 >= 4",
	}
	// remove date prefix on logs
	log.SetFlags(0)

	buffer := &bytes.Buffer{}
	streamLogger := NewStreamLog(buffer)

	for minLevel := NoLevel; minLevel <= ErrorLevel; minLevel++ {
		logger := NewVerbosityMiddleWare(minLevel, streamLogger)
		for logLevel := NoLevel; logLevel <= ErrorLevel; logLevel++ {
			logger.Logf(logLevel, "%d >= %d", logLevel, minLevel)
		}
	}
	logs := []string{}
	for line, err := buffer.ReadString('\n'); err == nil; line, err = buffer.ReadString('\n') {
		logs = append(logs, strings.Trim(line, "\n"))
	}
	assert.ElementsMatch(t, expected, logs)
	// clean up
	log.SetFlags(log.LstdFlags)
}
