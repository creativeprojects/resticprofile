//go:build !windows && !plan9

package main

import (
	"errors"
	"fmt"
	"log/syslog"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

type Syslog struct {
	writer *syslog.Writer
}

func NewSyslogHandler(writer *syslog.Writer) *Syslog {
	return &Syslog{
		writer: writer,
	}
}

func (l *Syslog) LogEntry(entry clog.LogEntry) error {
	if l.writer == nil {
		return errors.New("invalid syslog writer")
	}
	message := entry.GetMessage()
	switch entry.Level {
	case clog.LevelDebug:
		return l.writer.Debug(message)
	case clog.LevelInfo:
		return l.writer.Info(message)
	case clog.LevelWarning:
		return l.writer.Warning(message)
	case clog.LevelError:
		return l.writer.Err(message)
	default:
		return l.writer.Notice(message)
	}
}

func (l *Syslog) Close() error {
	err := l.writer.Close()
	l.writer = nil
	return err
}

var _ LogCloser = &Syslog{}

func getSyslogHandler(flags commandLineFlags, scheme, hostPort string) (*Syslog, error) {
	writer, err := syslog.Dial(scheme, hostPort, syslog.LOG_USER|syslog.LOG_NOTICE, constants.ApplicationName)
	if err != nil {
		return nil, fmt.Errorf("cannot open syslog logger: %w", err)
	}
	handler := NewSyslogHandler(writer)
	return handler, nil
}
