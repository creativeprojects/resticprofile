package config

import (
	"path/filepath"
	"strings"

	"github.com/creativeprojects/resticprofile/array"
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

var (
	// profileSections is a cache of ProfileSections()
	profileSections map[string][]string
)

// LoadConfiguration loads configuration from file
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

// SaveAs saves the current configuration into the file in parameter
func SaveAs(filename string) error {
	return viper.SafeWriteConfigAs(filename)
}

// ProfileKeys returns all profiles in the configuration
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

// ProfileGroups returns all groups from the configuration
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
	// Is the value cached?
	if profileSections != nil {
		return profileSections
	}
	allKeys := viper.AllKeys()
	if allKeys == nil || len(allKeys) == 0 {
		return nil
	}
	profileSections = make(map[string][]string, 0)
	for _, keys := range allKeys {
		keyPath := strings.SplitN(keys, ".", 3)
		if len(keyPath) > 0 {
			if keyPath[0] == constants.SectionConfigurationGlobal || keyPath[0] == constants.SectionConfigurationGroups {
				continue
			}
			var commands []string
			var found bool
			if commands, found = profileSections[keyPath[0]]; !found {
				commands = make([]string, 0)
			} else {
				commands = profileSections[keyPath[0]]
			}
			// If there's more than two keys, it means the second key is a group of keys, so it's a "command" definition
			if len(keyPath) > 2 {
				if _, found = array.FindString(commands, keyPath[1]); !found {
					commands = append(commands, keyPath[1])
				}
			}
			profileSections[keyPath[0]] = commands
		}
	}
	return profileSections
}
