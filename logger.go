package main

import (
	"io"
	"log"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/dial"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/creativeprojects/resticprofile/term"
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

func setupRemoteLogger(client *remote.Client) {
	logger := clog.NewLogger(client)
	client.SetPrefix("elevated user: ")
	clog.SetDefaultLogger(logger)
}

func setupTargetLogger(flags commandLineFlags) (io.Closer, error) {
	var (
		handler LogCloser
		file    *os.File
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

func getFileHandler(flags commandLineFlags) (*clog.StandardLogHandler, *os.File, error) {
	file, err := os.OpenFile(flags.log, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, err
	}
	return clog.NewStandardLogHandler(file, "", log.LstdFlags), file, nil
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
