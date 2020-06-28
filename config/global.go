package config

import (
	"github.com/creativeprojects/resticprofile/constants"
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

// newGlobal instantiates a new Global with default values
func newGlobal() *Global {
	return &Global{
		IONice:         constants.DefaultIONiceFlag,
		Nice:           constants.DefaultNiceFlag,
		DefaultCommand: constants.DefaultCommand,
		ResticBinary:   constants.DefaultResticBinary,
	}
}
