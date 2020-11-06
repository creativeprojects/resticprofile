package main

import (
	"log"
	"os"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/remote"
)

func setupRemoteLogger(client *remote.Client) {
	logger := clog.NewLogger(client)
	client.SetPrefix("elevated user: ")
	clog.SetDefaultLogger(logger)
}

func setupFileLogger(flags commandLineFlags) (*os.File, error) {
	file, err := os.OpenFile(flags.logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	logger := newFilteredLogger(flags, clog.NewStandardLogHandler(file, "", log.LstdFlags))
	// default logger added with level filtering
	clog.SetDefaultLogger(logger)
	// and return the file handle (so we can close it at the end)
	return file, nil
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

func newFilteredLogger(flags commandLineFlags, handler clog.Handler) *clog.Logger {
	if flags.quiet && (flags.verbose || flags.veryVerbose) {
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
			flags.veryVerbose = false
		}
		// the logger hasn't been created yet, so we call the handler directly
		handler.LogEntry(clog.LogEntry{
			Level:  clog.LevelWarning,
			Format: "you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.",
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
