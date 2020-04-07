package config

import (
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

// Global holds the configuration from the global section
type Global struct {
	IONice         bool   `mapstructure:"ionice"`
	IONiceClass    int    `mapstructure:"ionice-class"`
	IONiceLevel    int    `mapstructure:"ionice-level"`
	Nice           int    `mapstructure:"nice"`
	Priority       string `mapstructure:"priority"`
	DefaultCommand string `mapstructure:"default-command"`
	Initialize     bool   `mapstructure:"initialize"`
	ResticBinary   string `mapstructure:"restic-binary"`
}

// GetGlobalSection returns the global configuration
func GetGlobalSection() (*Global, error) {
	global := &Global{
		IONice:         constants.DefaultIONiceFlag,
		Nice:           constants.DefaultNiceFlag,
		DefaultCommand: constants.DefaultCommand,
		ResticBinary:   constants.DefaultResticBinary,
	}
	if viper.IsSet(constants.SectionConfigurationGlobal) {
		err := viper.UnmarshalKey(constants.SectionConfigurationGlobal, global)
		if err != nil {
			return nil, err
		}
	}
	return global, nil
}
