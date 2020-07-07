package clog

import (
	"io"
	"log"
	"os"
)

// LogLevel
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
)

// Log represents the logger interface
type Log interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warning(v ...interface{})
	Warningf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

// Verbosity represents the verbose logger interface
type Verbosity interface {
	Quiet()
	Verbose()
}

var (
	// default to null logger for tests
	defaultLog    Log       = &NullLog{}
	defaultOutput io.Writer = os.Stdout
)

func getLevelName(level int) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO "
	case WarningLevel:
		return "WARN "
	case ErrorLevel:
		return "ERROR"
	default:
		return ""
	}
}

// SetDefaultLogger sets the logger when using the package methods
func SetDefaultLogger(log Log) {
	defaultLog = log
}

// SetOutput sets the default output of the current logger
func SetOutput(w io.Writer) {
	defaultOutput = w
	log.SetOutput(w)
}

// GetOutput returns the default output of the current logger
func GetOutput() io.Writer {
	return defaultOutput
}

// Debug sends debugging information
func Debug(v ...interface{}) {
	defaultLog.Debug(v...)
}

// Debugf sends debugging information
func Debugf(format string, v ...interface{}) {
	defaultLog.Debugf(format, v...)
}

// Info logs some noticeable information
func Info(v ...interface{}) {
	defaultLog.Info(v...)
}

// Infof logs some noticeable information
func Infof(format string, v ...interface{}) {
	defaultLog.Infof(format, v...)
}

// Warning send some important message to the console
func Warning(v ...interface{}) {
	defaultLog.Warning(v...)
}

// Warningf send some important message to the console
func Warningf(format string, v ...interface{}) {
	defaultLog.Warningf(format, v...)
}

// Error sends error information to the console
func Error(v ...interface{}) {
	defaultLog.Error(v...)
}

// Errorf sends error information to the console
func Errorf(format string, v ...interface{}) {
	defaultLog.Errorf(format, v...)
}
