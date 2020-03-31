package clog

import (
	"log"
)

// LogLevel
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

var (
	quiet   bool
	verbose bool
	Levels  [4]*Color
)

func init() {
	Levels = [4]*Color{
		DebugLevel:   New(FgGreen),
		InfoLevel:    New(FgYellow),
		WarningLevel: New(FgRed),
		ErrorLevel:   New(FgRed),
	}
}

func SetLevel(quietFlag, verboseFlag bool) {
	quiet = quietFlag
	verbose = verboseFlag
}

func Debug(v ...interface{}) {
	if !verbose {
		return
	}
	message(Levels[DebugLevel], v...)
}

func Debugf(format string, v ...interface{}) {
	if !verbose {
		return
	}
	messagef(Levels[DebugLevel], format, v...)
}

func Info(v ...interface{}) {
	if quiet {
		return
	}
	message(Levels[InfoLevel], v...)
}

func Infof(format string, v ...interface{}) {
	if quiet {
		return
	}
	messagef(Levels[InfoLevel], format, v...)
}

func Warning(v ...interface{}) {
	message(Levels[WarningLevel], v...)
}

func Warningf(format string, v ...interface{}) {
	messagef(Levels[WarningLevel], format, v...)
}

func Error(v ...interface{}) {
	message(Levels[ErrorLevel], v...)
}

func Errorf(format string, v ...interface{}) {
	messagef(Levels[ErrorLevel], format, v...)
}

func message(color *Color, v ...interface{}) {
	color.Set()
	log.Println(v...)
	Unset()
}

func messagef(color *Color, format string, v ...interface{}) {
	color.Set()
	log.Printf(format+"\n", v...)
	Unset()
}
