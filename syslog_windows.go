//go:build windows

package main

import (
	"errors"
	"io"
)

func setupSyslogLogger(flags commandLineFlags) (io.Closer, error) {
	return nil, errors.New("syslog is not supported on Windows")
}
