package main

import (
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/remote"
)

func setupRemoteLogger(flags commandLineFlags) {
	client := remote.NewClient(flags.parentPort)
	logger := clog.NewRemoteLog(client)
	logger.SetPrefix("elevated user: ")
	clog.SetDefaultLogger(logger)
}

func setupFileLogger(flags commandLineFlags) (*clog.FileLog, error) {
	fileLogger, err := clog.NewFileLog(flags.logFile)
	if err != nil {
		return nil, err
	}
	logger := setupLevelMiddleware(flags, fileLogger)
	// default logger added with middleware
	clog.SetDefaultLogger(logger)
	// but return fileLogger (so we can close it at the end)
	return fileLogger, nil
}

func setupConsoleLogger(flags commandLineFlags) {
	consoleLogger := clog.NewConsoleLog()
	if flags.theme != "" {
		consoleLogger.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		consoleLogger.Colorize(false)
	}
	logger := setupLevelMiddleware(flags, consoleLogger)
	clog.SetDefaultLogger(logger)
}

func setupLevelMiddleware(flags commandLineFlags, logger clog.Logger) *clog.VerbosityMiddleware {
	if flags.quiet && flags.verbose {
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
		}
		logger.Warningf("you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.", coin)
	}
	minLevel := clog.InfoLevel
	if flags.quiet {
		minLevel = clog.WarningLevel
	} else if flags.verbose {
		minLevel = clog.NoLevel
	}
	return clog.NewVerbosityMiddleWare(minLevel, logger)
}
