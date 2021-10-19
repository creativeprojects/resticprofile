package config

import "strconv"

type ConfigVersion int

const (
	VersionUnknown ConfigVersion = iota
	Version01
	Version02
	VersionMax = Version02
)

// ParseVersion return the version number,
// if invalid the default version is Version01
func ParseVersion(raw string) ConfigVersion {
	if raw == "" {
		return Version01
	}
	version, err := strconv.Atoi(raw)
	if err != nil {
		return Version01
	}
	vers := ConfigVersion(version)
	if vers > VersionMax {
		return VersionMax
	}
	return vers
}
