package constants

import "time"

// Configuration defaults
const (
	DefaultConfigurationFile    = "profiles"
	DefaultProfileName          = "default"
	DefaultCommand              = "snapshots"
	DefaultFilterResticFlags    = true
	DefaultResticLockRetryAfter = 60 * time.Second
	DefaultResticStaleLockAge   = 2 * time.Hour
	DefaultTheme                = "light"
	DefaultIONiceFlag           = false
	DefaultIONiceClass          = 2
	DefaultStandardNiceFlag     = 0
	DefaultBackgroundNiceFlag   = 5
	DefaultVerboseFlag          = false
	DefaultQuietFlag            = false
	DefaultMinMemory            = 100
	DefaultSenderTimeout        = 30 * time.Second
)
