package config

import (
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/spf13/viper"
)

func LoadConfiguration(configFile string) error {
	viper.SetConfigType("toml")
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	used := viper.ConfigFileUsed()
	clog.Debugf("Found configuration file: %s", used)
	return nil
}
