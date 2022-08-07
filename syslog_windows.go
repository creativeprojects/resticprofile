//go:build windows

package main

import (
	"errors"
)

func getSyslogHandler(flags commandLineFlags, scheme, hostPort string) (LogCloser, error) {
	return nil, errors.New("syslog is not supported on Windows")
}
