package config

type ProfileUpgrader interface {
	Upgrade(key string, config *Config) error
}
