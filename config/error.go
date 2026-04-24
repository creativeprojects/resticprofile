package config

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrNotSupportedInVersion1 = errors.New("not supported in configuration version 1")
)
