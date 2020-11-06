package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config wraps up a viper configuration object
type Config struct {
	keyDelim       string
	format         string
	configFile     string
	viper          *viper.Viper
	groups         map[string][]string
	sourceTemplate *template.Template
}

// This is where things are getting hairy:
//
// Most configuration file formats allow only one declaration per section
// This is not the case for HCL where you can declare a bloc multiple times:
//
// "global" {
//   key1 = "value"
// }
//
// "global" {
//   key2 = "value"
// }
//
// For that matter, viper creates a slice of maps instead of a map for the other configuration file formats
// This configOptionHCL deals with the slice to merge it into a single map
var (
	configOption    = viper.DecodeHook(nil)
	configOptionHCL = viper.DecodeHook(sliceOfMapsToMapHookFunc())
)

// newConfig instantiate a new Config object
func newConfig(format string) *Config {
	return &Config{
		keyDelim: ".",
		format:   format,
		viper:    viper.New(),
	}
}

// LoadFile loads configuration from file
// Leave format blank for auto-detection from the file extension
func LoadFile(configFile, format string) (*Config, error) {
	if format == "" {
		// use file extension as format
		format = strings.TrimPrefix(filepath.Ext(configFile), ".")
	}
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot open configuration file for reading: %w", err)
	}
	c := newConfig(format)
	c.configFile = configFile
	// err = c.load(file)
	err = c.loadTemplate(file)
	if err != nil {
		return c, err
	}
	return c, nil
}

// Load configuration from reader
func Load(input io.Reader, format string) (*Config, error) {
	c := newConfig(format)
	// err := c.load(input)
	err := c.loadTemplate(input)
	if err != nil {
		return c, err
	}
	return c, nil
}

func (c *Config) loadTemplate(input io.Reader) error {
	inputString := &strings.Builder{}
	_, err := io.Copy(inputString, input)
	if err != nil {
		return err
	}
	c.sourceTemplate, err = template.New(filepath.Base(c.configFile)).Parse(inputString.String())
	if err != nil {
		return fmt.Errorf("cannot compile %w", err)
	}
	buffer := &bytes.Buffer{}
	err = c.sourceTemplate.Execute(buffer, newTemplateData(c.configFile, "default"))
	if err != nil {
		return fmt.Errorf("cannot execute %w", err)
	}
	traceConfig("default", buffer.String())
	return c.load(buffer)
}

func (c *Config) load(input io.Reader) error {
	// For compatibility with the previous versions, a .conf file is TOML format
	if c.format == "conf" {
		c.format = "toml"
	}
	c.viper.SetConfigType(c.format)
	err := c.viper.ReadConfig(input)
	if err != nil {
		return fmt.Errorf("cannot parse %s configuration: %w", c.format, err)
	}
	return nil
}

func (c *Config) reloadTemplate(data TemplateData) error {
	if c.sourceTemplate == nil {
		return errors.New("no available template to execute, please load it first")
	}
	buffer := &bytes.Buffer{}
	err := c.sourceTemplate.Execute(buffer, data)
	if err != nil {
		return fmt.Errorf("cannot execute %w", err)
	}
	traceConfig(data.Profile.Name, buffer.String())
	return c.load(buffer)
}

// IsSet checks if the key contains a value
func (c *Config) IsSet(key string) bool {
	if strings.Contains(key, ".") {
		clog.Warningf("it should not search for a subkey: %s", key)
	}
	return c.viper.IsSet(key)
}

// GetConfigFile returns the config file used
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// Get the value from the key
func (c *Config) Get(key string) interface{} {
	return c.viper.Get(key)
}

// HasProfile returns true if the profile exists in the configuration
func (c *Config) HasProfile(profileKey string) bool {
	return c.IsSet(profileKey)
}

// AllSettings merges all settings and returns them as a map[string]interface{}.
func (c *Config) AllSettings() map[string]interface{} {
	return c.viper.AllSettings()
}

// GetProfileSections returns a list of profiles with all the sections defined inside each
func (c *Config) GetProfileSections() map[string][]string {
	profiles := map[string][]string{}
	allSettings := c.AllSettings()
	for sectionKey, sectionRawValue := range allSettings {
		if sectionKey == constants.SectionConfigurationGlobal || sectionKey == constants.SectionConfigurationGroups {
			continue
		}
		var commandList []string
		if c.format == "hcl" {
			commandList = c.getCommandListHCL(sectionRawValue)
		} else {
			commandList = c.getCommandList(sectionRawValue)
		}
		profiles[sectionKey] = commandList
	}
	return profiles
}

func (c *Config) getCommandList(sectionRawValue interface{}) []string {
	commandList := []string{}
	if sectionValues, ok := sectionRawValue.(map[string]interface{}); ok {
		// For each value in here, if it's a map it means it's defining some command parameters
		for key, value := range sectionValues {
			if _, ok := value.(map[string]interface{}); ok {
				commandList = append(commandList, key)
			}
		}
	}
	return commandList
}

func (c *Config) getCommandListHCL(sectionRawValue interface{}) []string {
	commandList := []string{}
	if sectionValues, ok := sectionRawValue.([]map[string]interface{}); ok {
		// for each map in the array
		for _, subMap := range sectionValues {
			// for each value in here, if it's a map it means it's defining some command parameters
			for key, value := range subMap {
				// Special case for hcl where each map will be wrapped around a list
				if _, ok := value.([]map[string]interface{}); ok {
					commandList = append(commandList, key)
				}
			}
		}
	}
	return commandList
}

// GetGlobalSection returns the global configuration
func (c *Config) GetGlobalSection() (*Global, error) {
	global := newGlobal()
	err := c.unmarshalKey(constants.SectionConfigurationGlobal, global)
	if err != nil {
		return nil, err
	}
	return global, nil
}

// HasProfileGroup returns true if the group of profiles exists in the configuration
func (c *Config) HasProfileGroup(groupKey string) bool {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return false
	}
	err := c.loadGroups()
	if err != nil {
		return false
	}
	_, ok := c.groups[groupKey]
	return ok
}

// GetProfileGroup returns the list of profiles in a group
func (c *Config) GetProfileGroup(groupKey string) ([]string, error) {
	err := c.loadGroups()
	if err != nil {
		return nil, err
	}

	group, ok := c.groups[groupKey]
	if !ok {
		return nil, fmt.Errorf("group '%s' not found", groupKey)
	}
	return group, nil
}

// GetProfileGroups returns all groups from the configuration
//
// If the groups section does not exist, it returns an empty map
func (c *Config) GetProfileGroups() map[string][]string {
	err := c.loadGroups()
	if err != nil {
		return nil
	}
	return c.groups
}

func (c *Config) loadGroups() error {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		c.groups = map[string][]string{}
		return nil
	}
	if c.groups == nil {
		groups := map[string][]string{}
		err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
		if err != nil {
			return err
		}
		c.groups = groups
	}
	return nil
}

// GetProfile in configuration
func (c *Config) GetProfile(profileKey string) (*Profile, error) {
	if c.sourceTemplate != nil {
		err := c.reloadTemplate(newTemplateData(c.configFile, profileKey))
		if err != nil {
			return nil, err
		}
	}
	return c.getProfile(profileKey)
}

// getProfile from configuration
func (c *Config) getProfile(profileKey string) (*Profile, error) {
	var err error
	var profile *Profile

	if !c.IsSet(profileKey) {
		return nil, nil
	}

	profile = NewProfile(c, profileKey)
	err = c.unmarshalKey(profileKey, profile)
	if err != nil {
		return nil, err
	}
	if profile.Inherit != "" {
		inherit := profile.Inherit
		// Load inherited profile
		profile, err = c.getProfile(inherit)
		if err != nil {
			return nil, err
		}
		if profile == nil {
			return nil, fmt.Errorf("error in profile '%s': parent profile '%s' not found", profileKey, inherit)
		}
		// and reload this profile onto the inherited one
		err = c.unmarshalKey(profileKey, profile)
		if err != nil {
			return nil, err
		}
		// make sure it has the right name
		profile.Name = profileKey
	}
	return profile, nil
}

// unmarshalKey is a wrapper around viper.UnmarshalKey with the right decoder config options
func (c *Config) unmarshalKey(key string, rawVal interface{}) error {
	if c.format == "hcl" {
		return c.viper.UnmarshalKey(key, rawVal, configOptionHCL)
	}
	return c.viper.UnmarshalKey(key, rawVal, configOption)
}

// sliceOfMapsToMapHookFunc merges a slice of maps to a map
func sliceOfMapsToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Slice && from.Elem().Kind() == reflect.Map && (to.Kind() == reflect.Struct || to.Kind() == reflect.Map) {
			// clog.Debugf("hook: from slice %+v to %+v", from.Elem(), to)
			source, ok := data.([]map[string]interface{})
			if !ok {
				return data, nil
			}
			if len(source) == 0 {
				return data, nil
			}
			if len(source) == 1 {
				return source[0], nil
			}
			// flatten the slice into one map
			convert := make(map[string]interface{})
			for _, mapItem := range source {
				for key, value := range mapItem {
					convert[key] = value
				}
			}
			return convert, nil
		}
		// clog.Debugf("default from %+v to %+v", from, to)
		return data, nil
	}
}

func traceConfig(profileName, config string) {
	lines := strings.Split(config, "\n")
	output := ""
	for i := 0; i < len(lines); i++ {
		output += fmt.Sprintf("%3d: %s\n", i+1, lines[i])
	}
	clog.Tracef("Resulting configuration for profile '%s':\n====================\n%s====================\n", profileName, output)
}
