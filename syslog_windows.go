//go:build windows

package main

import (
	"errors"
	"io"
)

func getSyslogHandler(scheme, hostPort string) (_ LogCloser, _ io.Writer, err error) {
	err = errors.New("syslog is not supported on Windows")
	return
}
