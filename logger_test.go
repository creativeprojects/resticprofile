package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readTail(t *testing.T, filename string, count int) (lines []string) {
	t.Helper()

	file, e := os.Open(filename)
	require.NoError(t, e)
	defer file.Close()
	for s := bufio.NewScanner(file); s.Scan(); require.NoError(t, s.Err()) {
		lines = append(lines, s.Text())
	}
	if len(lines) < count {
		count = len(lines)
	}
	return lines[len(lines)-count:]
}

func TestFileHandlerWithTemporaryDirMarker(t *testing.T) {
	util.ClearTempDir()
	defer util.ClearTempDir()

	logFile := filepath.Join(util.MustGetTempDir(), "sub", "file.log")
	assert.NoFileExists(t, logFile)

	handler, _, err := getFileHandler(filepath.Join(constants.TemporaryDirMarker, "sub", "file.log"))
	require.NoError(t, err)
	assert.FileExists(t, logFile)

	assert.NoError(t, handler.Close())
	util.ClearTempDir()
	assert.NoFileExists(t, logFile)
}

func TestFileHandler(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "file.log")
	handler, writer, err := getFileHandler(logFile)
	require.NoError(t, err)
	defer handler.Close()

	require.Implements(t, (*term.Flusher)(nil), writer)
	flusher := writer.(term.Flusher)

	log := func(line string) {
		assert.NoError(t, handler.LogEntry(clog.LogEntry{Level: clog.LevelInfo, Format: line}))
	}

	// output is accessible while handler is not closed
	{
		log("log-line-1")
		assert.NoError(t, flusher.Flush())

		lines := readTail(t, logFile, 1)
		require.NotEmpty(t, lines)
		assert.Regexp(t, `^.+\slog-line-1$`, lines[0])
	}

	// output is buffered
	{
		log("log-line-2")
		assert.Regexp(t, `^.+\slog-line-1$`, readTail(t, logFile, 1)[0])

		assert.NoError(t, flusher.Flush())
		assert.Regexp(t, `^.+\slog-line-2$`, readTail(t, logFile, 1)[0])
	}

	// output is auto-flushed
	{
		log("log-line-3")
		assert.Regexp(t, `^.+\slog-line-2$`, readTail(t, logFile, 1)[0])

		time.Sleep(300 * time.Millisecond)
		assert.Regexp(t, `^.+\slog-line-3$`, readTail(t, logFile, 1)[0])
	}

	// output is formatted as expected
	{
		lines := readTail(t, logFile, 10)
		require.Len(t, lines, 3)
		for i := 0; i < 3; i++ {
			assert.Regexp(t, fmt.Sprintf(`^.+\slog-line-%d$`, i+1), lines[i])
		}
	}
}

func TestParseCommandOutput(t *testing.T) {
	tests := []struct {
		co       string
		all, log bool
	}{
		{co: "", all: false, log: false},
		{co: "auto", all: term.OsStdoutIsTerminal(), log: true},
		{co: "log", all: false, log: true},
		{co: "console", all: false, log: false},
		{co: "all", all: true, log: false},
		{co: "all,log", all: true, log: true},
		{co: "console,log", all: true, log: true},
		{co: "log,console", all: true, log: true},
		{co: "log,a", all: false, log: true},
		{co: "console,a", all: false, log: false},

		{co: " auto ", all: term.OsStdoutIsTerminal(), log: true},
		{co: " all ", all: true, log: false},
		{co: " log ", all: false, log: true},
		{co: " console ", all: false, log: false},
		{co: " console , log ", all: true, log: true},
		{co: " log , console ", all: true, log: true},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			a, l := parseCommandOutput(test.co)
			assert.Equal(t, test.all, a, "all")
			assert.Equal(t, test.log, l, "log")
		})
	}
}

func TestCloseFileHandler(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "file.log")
	handler, writer, err := getFileHandler(logFile)
	require.NoError(t, err)
	assert.NotNil(t, handler)
	assert.NotNil(t, writer)

	assert.NoError(t, handler.LogEntry(clog.LogEntry{Level: clog.LevelInfo, Format: "log-line-1"}))
	handler.Close()
	assert.Error(t, handler.LogEntry(clog.LogEntry{Level: clog.LevelInfo, Format: "log-line-2"}))
}
