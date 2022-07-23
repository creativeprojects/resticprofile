//go:build !windows && !plan9

package main

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net/url"

	"github.com/creativeprojects/clog"
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

var _ clog.Handler = &Syslog{}

func setupSyslogLogger(flags commandLineFlags) (io.Closer, error) {
	scheme, hostPort, err := getDialAddr(flags.syslog)
	if err != nil {
		return nil, err
	}
	writer, err := syslog.Dial(scheme, hostPort, syslog.LOG_USER|syslog.LOG_NOTICE, "resticprofile")
	if err != nil {
		return nil, fmt.Errorf("cannot open syslog logger: %w", err)
	}
	handler := NewSyslogHandler(writer)
	// use the console handler as a backup
	logger := newFilteredLogger(flags, clog.NewSafeHandler(handler, clog.NewConsoleHandler("", log.LstdFlags)))
	clog.SetDefaultLogger(logger)
	return writer, nil
}

func getDialAddr(source string) (string, string, error) {
	URL, err := url.Parse(source)
	if err != nil {
		return "", "", err
	}
	scheme := URL.Scheme
	hostPort := URL.Host
	return scheme, hostPort, nil
}
