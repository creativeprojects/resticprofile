package config

import (
	"strings"

	"github.com/creativeprojects/resticprofile/constants"

	"github.com/creativeprojects/resticprofile/array"
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

func ProfileKeys() []string {
	allKeys := viper.AllKeys()
	if allKeys == nil || len(allKeys) == 0 {
		return nil
	}
	profiles := make([]string, 0)
	for _, keys := range allKeys {
		keyPath := strings.SplitN(keys, ".", 2)
		if len(keyPath) > 0 {
			if keyPath[0] == constants.SectionConfigurationGlobal || keyPath[0] == constants.SectionConfigurationGroups {
				continue
			}
			if _, found := array.FindString(profiles, keyPath[0]); !found {
				profiles = append(profiles, keyPath[0])
			}
		}
	}
	return profiles
}
