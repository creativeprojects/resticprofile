package batt

import "errors"

var ErrBatteryInfoNotSupported = errors.New("battery information is not supported on this platform")
