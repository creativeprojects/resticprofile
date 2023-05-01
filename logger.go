package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/dial"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/creativeprojects/resticprofile/util"
)

type LogCloser interface {
	clog.Handler
	Close() error
}

func setupConsoleLogger(flags commandLineFlags) {
	consoleHandler := clog.NewConsoleHandler("", log.LstdFlags)
	if flags.theme != "" {
		consoleHandler.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		consoleHandler.Colouring(false)
	}
	logger := newFilteredLogger(flags, consoleHandler)
	clog.SetDefaultLogger(logger)
}

func setupRemoteLogger(flags commandLineFlags, client *remote.Client) {
	client.SetPrefix("elevated user: ")
	logger := newFilteredLogger(flags, clog.NewLogger(client))
	clog.SetDefaultLogger(logger)
}

func setupTargetLogger(flags commandLineFlags) (io.Closer, error) {
	var (
		handler LogCloser
		file    io.Writer
		err     error
	)
	scheme, hostPort, isURL := dial.GetAddr(flags.log)
	if isURL {
		handler, err = getSyslogHandler(flags, scheme, hostPort)
	} else {
		handler, file, err = getFileHandler(flags)
	}
	if err != nil {
		return nil, err
	}
	// use the console handler as a backup
	logger := newFilteredLogger(flags, clog.NewSafeHandler(handler, clog.NewConsoleHandler("", log.LstdFlags)))
	// default logger added with level filtering
	clog.SetDefaultLogger(logger)

	// also redirect all terminal output
	if file != nil {
		term.SetAllOutput(file)
	}
	// and return the handler (so we can close it at the end)
	return handler, nil
}

func getFileHandler(flags commandLineFlags) (*clog.StandardLogHandler, io.Writer, error) {
	if strings.HasPrefix(flags.log, constants.TemporaryDirMarker) {
		if tempDir, err := util.TempDir(); err == nil {
			flags.log = flags.log[len(constants.TemporaryDirMarker):]
			if len(flags.log) > 0 && os.IsPathSeparator(flags.log[0]) {
				flags.log = flags.log[1:]
			}
			flags.log = filepath.Join(tempDir, flags.log)
			_ = os.MkdirAll(filepath.Dir(flags.log), 0755)
		}
	}

	// create a platform aware log file appender
	var appender util.AsyncFileWriterAppendFunc
	if platform.IsWindows() {
		appender = func(dst []byte, c byte) []byte {
			switch c {
			case '\n':
				return append(dst, '\r', '\n') // normalize to CRLF on Windows
			case '\r':
				return dst
			}
			return append(dst, c)
		}
	}

	writer, err := util.NewAsyncFileWriter(
		flags.log,
		util.WithAsyncFileAppendFunc(appender),
		util.WithAsyncFilePerm(0644),
	)
	if err != nil {
		return nil, nil, err
	}

	return clog.NewStandardLogHandler(writer, "", log.LstdFlags), writer, nil
}

func newFilteredLogger(flags commandLineFlags, handler clog.Handler) *clog.Logger {
	if flags.quiet && (flags.verbose || flags.veryVerbose) {
		var coin string
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
			flags.veryVerbose = false
		}
		// the logger hasn't been created yet, so we call the handler directly
		_ = handler.LogEntry(clog.LogEntry{
			Level:  clog.LevelWarning,
			Format: "you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! ... and the winner is ... %s.",
			Values: []interface{}{coin},
		})
	}
	minLevel := clog.LevelInfo
	if flags.quiet {
		minLevel = clog.LevelWarning
	} else if flags.veryVerbose {
		minLevel = clog.LevelTrace
	} else if flags.verbose {
		minLevel = clog.LevelDebug
	}
	// now create and return the logger
	return clog.NewLogger(clog.NewLevelFilter(minLevel, handler))
}

func changeLevelFilter(level clog.LogLevel) {
	handler := clog.GetDefaultLogger().GetHandler()
	filter, ok := handler.(*clog.LevelFilter)
	if ok {
		filter.SetLevel(level)
	}
}
