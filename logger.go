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
	logger, err := clog.NewFileLog(flags.logFile)
	if err != nil {
		return nil, err
	}
	clog.SetDefaultLogger(logger)
	setupVerbosity(flags, logger)
	return logger, nil
}

func setupConsoleLogger(flags commandLineFlags) {
	logger := clog.NewConsoleLog()
	if flags.theme != "" {
		logger.SetTheme(flags.theme)
	}
	if flags.noAnsi {
		logger.Colorize(false)
	}
	clog.SetDefaultLogger(logger)
	setupVerbosity(flags, logger)
}

func setupVerbosity(flags commandLineFlags, logger clog.Verbosity) {
	if flags.quiet && flags.verbose {
		coin := ""
		if randomBool() {
			coin = "verbose"
			flags.quiet = false
		} else {
			coin = "quiet"
			flags.verbose = false
		}
		clog.Warningf("you specified -quiet (-q) and -verbose (-v) at the same time. So let's flip a coin! and selection is ... %s.", coin)
	}
	if flags.quiet {
		logger.Quiet()
	}
	if flags.verbose {
		logger.Verbose()
	}
}
