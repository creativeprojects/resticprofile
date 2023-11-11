//go:build windows

package main

import (
	"errors"
)

func getSyslogHandler(scheme, hostPort string) (LogCloser, error) {
	return nil, errors.New("syslog is not supported on Windows")
}
