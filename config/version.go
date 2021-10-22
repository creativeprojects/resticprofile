package config

import "strconv"

type Version int

const (
	VersionUnknown Version = iota
	Version01
	Version02
	VersionMax = Version02
)

// ParseVersion return the version number,
// if invalid the default version is Version01
func ParseVersion(raw string) Version {
	if raw == "" {
		return Version01
	}
	version, err := strconv.Atoi(raw)
	if err != nil {
		return Version01
	}
	vers := Version(version)
	if vers > VersionMax {
		return VersionMax
	}
	return vers
}
