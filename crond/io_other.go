//go:build !linux

package crond

// DefaultCrontabBinary is empty as this platform has no default `crontab` executable
const DefaultCrontabBinary = ""
