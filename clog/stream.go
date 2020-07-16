package clog

import (
	"io"
	"log"
)

// StreamLog logs messages to a stream implementing the io.Writer interface
type StreamLog struct {
	writer io.Writer
}

// NewStreamLog creates a new stream logger
func NewStreamLog(w io.Writer) *StreamLog {
	logger := &StreamLog{
		writer: w,
	}
	// also redirect the standard logger to the stream
	SetOutput(w)
	return logger
}

// Log sends a log entry with the specified level
func (l *StreamLog) Log(level LogLevel, v ...interface{}) {
	l.message(level, v...)
}

// Logf sends a log entry with the specified level
func (l *StreamLog) Logf(level LogLevel, format string, v ...interface{}) {
	l.messagef(level, format, v...)
}

// Debug sends debugging information
func (l *StreamLog) Debug(v ...interface{}) {
	l.message(LevelDebug, v...)
}

// Debugf sends debugging information
func (l *StreamLog) Debugf(format string, v ...interface{}) {
	l.messagef(LevelDebug, format, v...)
}

// Info logs some noticeable information
func (l *StreamLog) Info(v ...interface{}) {
	l.message(LevelInfo, v...)
}

// Infof logs some noticeable information
func (l *StreamLog) Infof(format string, v ...interface{}) {
	l.messagef(LevelInfo, format, v...)
}

// Warning send some important message to the console
func (l *StreamLog) Warning(v ...interface{}) {
	l.message(LevelWarning, v...)
}

// Warningf send some important message to the console
func (l *StreamLog) Warningf(format string, v ...interface{}) {
	l.messagef(LevelWarning, format, v...)
}

// Error sends error information to the console
func (l *StreamLog) Error(v ...interface{}) {
	l.message(LevelError, v...)
}

// Errorf sends error information to the console
func (l *StreamLog) Errorf(format string, v ...interface{}) {
	l.messagef(LevelError, format, v...)
}

func (l *StreamLog) message(level LogLevel, v ...interface{}) {
	v = append([]interface{}{level.String()}, v...)
	log.Println(v...)
}

func (l *StreamLog) messagef(level LogLevel, format string, v ...interface{}) {
	log.Printf(level.String()+" "+format+"\n", v...)
}

// Verify interface
var (
	_ Logger = &StreamLog{}
)
