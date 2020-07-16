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

	Log(LevelInfo, "one", "two", "three")
	Debug("one", "two", "three")
	Info("one", "two", "three")
	Warning("one", "two", "three")
	Error("one", "two", "three")

	Logf(LevelInfo, "%d %d %d", 1, 2, 3)
	Debugf("%d %d %d", 1, 2, 3)
	Infof("%d %d %d", 1, 2, 3)
	Warningf("%d %d %d", 1, 2, 3)
	Errorf("%d %d %d", 1, 2, 3)
}

func TestFileLoggerConcurrency(t *testing.T) {
	// remove date prefix on logs during the test
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)

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
}

func TestLoggerVerbosity(t *testing.T) {
	expected := []string{
		"DEBUG 0 >= 0",
		"INFO  1 >= 0",
		"WARN  2 >= 0",
		"ERROR 3 >= 0",
		"INFO  1 >= 1",
		"WARN  2 >= 1",
		"ERROR 3 >= 1",
		"WARN  2 >= 2",
		"ERROR 3 >= 2",
		"ERROR 3 >= 3",
	}
	// remove date prefix on logs during the test
	log.SetFlags(0)
	defer log.SetFlags(log.LstdFlags)

	buffer := &bytes.Buffer{}
	streamLogger := NewStreamLog(buffer)

	for minLevel := LevelDebug; minLevel <= LevelError; minLevel++ {
		logger := NewLevelFilter(minLevel, streamLogger)
		for logLevel := LevelDebug; logLevel <= LevelError; logLevel++ {
			logger.Logf(logLevel, "%d >= %d", logLevel, minLevel)
		}
	}
	logs := []string{}
	for line, err := buffer.ReadString('\n'); err == nil; line, err = buffer.ReadString('\n') {
		logs = append(logs, strings.Trim(line, "\n"))
	}
	assert.ElementsMatch(t, expected, logs)
}

func BenchmarkStreamMessages(b *testing.B) {
	b.ReportAllocs()
	buffer := &bytes.Buffer{}
	streamLogger := NewStreamLog(buffer)
	logger := NewLevelFilter(LevelDebug, streamLogger)
	param1 := "string"
	param2 := 0

	for i := 0; i < b.N; i++ {
		logger.Infof("Message with a %s and a %d", param1, param2)
	}
}

func BenchmarkStreamFilteredMessages(b *testing.B) {
	b.ReportAllocs()
	buffer := &bytes.Buffer{}
	streamLogger := NewStreamLog(buffer)
	logger := NewLevelFilter(LevelWarning, streamLogger)
	param1 := "string"
	param2 := 0

	for i := 0; i < b.N; i++ {
		logger.Infof("Message with a %s and a %d", param1, param2)
	}
}
