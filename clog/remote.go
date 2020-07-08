package clog

import (
	"fmt"
	"log"
)

// RemoteLogClient represents the interface of a remote logger
type RemoteLogClient interface {
	// Log with a level from 0 to 4
	Log(level int, message string) error
}

// RemoteLog logs messages to a remote logger
type RemoteLog struct {
	prefix string
	client RemoteLogClient
}

// NewRemoteLog creates a new remote logger
func NewRemoteLog(client RemoteLogClient) *RemoteLog {
	console := &RemoteLog{
		client: client,
	}
	return console
}

// SetPrefix on all messages sent to the server
func (l *RemoteLog) SetPrefix(prefix string) {
	l.prefix = prefix
}

// Log sends a log entry with the specified level
func (l *RemoteLog) Log(level LogLevel, v ...interface{}) {
	l.message(level, v...)
}

// Logf sends a log entry with the specified level
func (l *RemoteLog) Logf(level LogLevel, format string, v ...interface{}) {
	l.messagef(level, format, v...)
}

// Debug sends debugging information
func (l *RemoteLog) Debug(v ...interface{}) {
	l.message(DebugLevel, v...)
}

// Debugf sends debugging information
func (l *RemoteLog) Debugf(format string, v ...interface{}) {
	l.messagef(DebugLevel, format, v...)
}

// Info logs some noticeable information
func (l *RemoteLog) Info(v ...interface{}) {
	l.message(InfoLevel, v...)
}

// Infof logs some noticeable information
func (l *RemoteLog) Infof(format string, v ...interface{}) {
	l.messagef(InfoLevel, format, v...)
}

// Warning send some important message to the console
func (l *RemoteLog) Warning(v ...interface{}) {
	l.message(WarningLevel, v...)
}

// Warningf send some important message to the console
func (l *RemoteLog) Warningf(format string, v ...interface{}) {
	l.messagef(WarningLevel, format, v...)
}

// Error sends error information to the console
func (l *RemoteLog) Error(v ...interface{}) {
	l.message(ErrorLevel, v...)
}

// Errorf sends error information to the console
func (l *RemoteLog) Errorf(format string, v ...interface{}) {
	l.messagef(ErrorLevel, format, v...)
}

func (l *RemoteLog) message(level LogLevel, v ...interface{}) {
	v = append([]interface{}{l.prefix}, v...)
	err := l.client.Log(int(level), fmt.Sprint(v...))
	if err != nil {
		log.Println(err)
	}
}

func (l *RemoteLog) messagef(level LogLevel, format string, v ...interface{}) {
	err := l.client.Log(int(level), fmt.Sprintf(l.prefix+format, v...))
	if err != nil {
		log.Println(err)
	}
}

// Verify interface
var (
	_ Logger = &RemoteLog{}
)
