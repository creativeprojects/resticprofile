package config

import "strconv"

type ConfigVersion int

const (
	Version00 ConfigVersion = iota
	Version01
	Version02
	VersionMax = Version02
)

func ParseVersion(raw string) ConfigVersion {
	version, err := strconv.Atoi(raw)
	if err != nil {
		return Version00
	}
	vers := ConfigVersion(version)
	if vers > VersionMax {
		return VersionMax
	}
	return vers
}
