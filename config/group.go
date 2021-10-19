package config

// Group of profiles
type Group struct {
	Description string   `mapstructure:"description"`
	Profiles    []string `mapstructure:"profiles"`
}
