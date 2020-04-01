package config

import (
	"path/filepath"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"

	"github.com/creativeprojects/resticprofile/array"
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/spf13/viper"
)

func LoadConfiguration(configFile string) error {
	// For compatibility with the previous versions, a .conf file is TOML format
	if filepath.Ext(configFile) == ".conf" {
		viper.SetConfigType("toml")
	}
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

func ProfileGroups() map[string][]string {
	groups := make(map[string][]string, 0)
	if !viper.IsSet(constants.SectionConfigurationGroups) {
		return nil
	}
	err := viper.UnmarshalKey(constants.SectionConfigurationGroups, &groups)
	if err != nil {
		return nil
	}
	return groups
}

// ProfileSections returns a list of profiles with all the sections defined inside each
func ProfileSections() map[string][]string {
	allKeys := viper.AllKeys()
	if allKeys == nil || len(allKeys) == 0 {
		return nil
	}
	profiles := make(map[string][]string, 0)
	for _, keys := range allKeys {
		keyPath := strings.SplitN(keys, ".", 3)
		if len(keyPath) > 0 {
			if keyPath[0] == constants.SectionConfigurationGlobal || keyPath[0] == constants.SectionConfigurationGroups {
				continue
			}
			var commands []string
			var found bool
			if commands, found = profiles[keyPath[0]]; !found {
				commands = make([]string, 0)
			} else {
				commands = profiles[keyPath[0]]
			}
			// If there's more than two keys, it means the second key is a group of keys, so it's a "command" definition
			if len(keyPath) > 2 {
				if _, found = array.FindString(commands, keyPath[1]); !found {
					commands = append(commands, keyPath[1])
				}
			}
			profiles[keyPath[0]] = commands
		}
	}
	return profiles
}
