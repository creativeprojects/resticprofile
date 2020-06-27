package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/resticprofile/array"
	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

// Config wraps up a viper configuration object
type Config struct {
	keyDelim string
	format   string
	viper    *viper.Viper
	data     map[string]interface{}
}

// NewConfig instantiate a new Config object
func NewConfig() *Config {
	return &Config{
		keyDelim: ".",
		viper:    viper.New(),
		data:     make(map[string]interface{}),
	}
}

// LoadFile loads configuration from file
func (c *Config) LoadFile(configFile string) error {
	format := filepath.Ext(configFile)
	if strings.HasPrefix(format, ".") {
		format = format[1:]
	}
	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("cannot open configuration file for reading: %w", err)
	}
	return c.Load(file, format)
}

// Load configuration from reader
func (c *Config) Load(input io.Reader, format string) error {
	// For compatibility with the previous versions, a .conf file is TOML format
	if format == "conf" {
		format = "toml"
	}
	c.format = format
	c.viper.SetConfigType(c.format)
	return c.viper.ReadConfig(input)
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are separated by a "."
func (c *Config) AllKeys() []string {
	return c.viper.AllKeys()
}

// IsSet checks if the key contains a value
func (c *Config) IsSet(key string) bool {
	if strings.Contains(key, ".") {
		clog.Warningf("it should not search for a subkey: %s", key)
	}
	return c.viper.IsSet(key)
}

// Get the value from the key
func (c *Config) Get(key string) interface{} {
	return c.viper.Get(key)
}

// HasProfile returns true if the profile exists in the configuration
func (c *Config) HasProfile(profileKey string) bool {
	return c.IsSet(profileKey)
}

// ProfileGroups returns all groups from the configuration
func (c *Config) ProfileGroups() map[string][]string {
	groups := make(map[string][]string, 0)
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return nil
	}
	err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
	if err != nil {
		return nil
	}
	return groups
}

// ProfileSections returns a list of profiles with all the sections defined inside each
func (c *Config) ProfileSections() map[string][]string {
	allKeys := c.AllKeys()
	if allKeys == nil || len(allKeys) == 0 {
		return nil
	}
	profileSections := make(map[string][]string, 0)
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
			// If there are more than two keys, it means the second key is a group of keys, so it's a "command" definition
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

// SaveAs saves the current configuration into the file in parameter
func (c *Config) SaveAs(filename string) error {
	return c.viper.SafeWriteConfigAs(filename)
}

// unmarshalKey is a wrapper around viper.UnmarshalKey with default decoder config options
func (c *Config) unmarshalKey(key string, rawVal interface{}) error {
	return c.viper.UnmarshalKey(key, rawVal, configOption)
}
