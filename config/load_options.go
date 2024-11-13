package config

func WithConfigFile(configFile string) func(cfg *Config) {
	return func(cfg *Config) {
		cfg.configFile = configFile
	}
}
