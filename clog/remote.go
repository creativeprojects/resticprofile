package clog

import (
	"fmt"
	"log"
)

// RemoteLogClient represents the interface of a remote logger
type RemoteLogClient interface {
	Log(level int, message string) error
}

// RemoteLog logs messages to the console (in colour)
type RemoteLog struct {
	prefix string
	client RemoteLogClient
}

// NewRemoteLog creates a new console logger
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

func (l *RemoteLog) message(level int, v ...interface{}) {
	v = append([]interface{}{l.prefix}, v...)
	err := l.client.Log(level, fmt.Sprint(v...))
	if err != nil {
		log.Println(err)
	}
}

func (l *RemoteLog) messagef(level int, format string, v ...interface{}) {
	err := l.client.Log(level, fmt.Sprintf(l.prefix+format, v...))
	if err != nil {
		log.Println(err)
	}
}

// Verify interface
var (
	_ Log = &RemoteLog{}
)
