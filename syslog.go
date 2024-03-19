//go:build !windows && !plan9

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"net"
	"strings"

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

const DefaultSyslogPort = "514"

type tokenWriter struct {
	separator []byte
	target    io.Writer
}

func (s *tokenWriter) Write(p []byte) (n int, err error) {
	var pn int
	for i, part := range bytes.Split(p, s.separator) {
		if err != nil {
			break
		}
		pn, err = s.target.Write(part)
		n += pn
		if i > 0 {
			n += len(s.separator)
		}
	}
	return
}

func getSyslogHandler(scheme, hostPort string) (handler *Syslog, writer io.Writer, err error) {
	switch scheme {
	case "udp", "tcp":
	case "syslog-tcp":
		scheme = "tcp"
	case "syslog":
		if len(hostPort) == 0 {
			scheme = "local"
		} else {
			scheme = "udp"
		}
	default:
		err = fmt.Errorf("unsupported syslog URL scheme %q", scheme)
		return
	}

	var logger *syslog.Writer
	if scheme == "local" {
		logger, err = syslog.New(syslog.LOG_USER|syslog.LOG_NOTICE, constants.ApplicationName)
	} else {
		if _, _, e := net.SplitHostPort(hostPort); e != nil && strings.Contains(e.Error(), "missing port") {
			hostPort = net.JoinHostPort(hostPort, DefaultSyslogPort)
		}
		logger, err = syslog.Dial(scheme, hostPort, syslog.LOG_USER|syslog.LOG_NOTICE, constants.ApplicationName)
	}

	if err == nil {
		writer = &tokenWriter{separator: []byte("\n"), target: logger}
		handler = NewSyslogHandler(logger)
	} else {
		err = fmt.Errorf("cannot open syslog logger: %w", err)
	}
	return
}
