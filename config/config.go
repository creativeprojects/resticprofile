package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config wraps up a viper configuration object
type Config struct {
	keyDelim        string
	format          string
	configFile      string
	includeFiles    []string
	viper           *viper.Viper
	groups          map[string]Group
	sourceTemplates *template.Template
	version         Version
	issues          struct {
		changedPaths map[string][]string // 'path' items that had been changed to absolute paths
	}
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
	configOption = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		confidentialValueDecoder(),
	))

	configOptionHCL = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		confidentialValueDecoder(),
		sliceOfMapsToMapHookFunc(),
	))

	rootPathMessage = sync.Once{}
)

// newConfig instantiate a new Config object
func newConfig(format string) *Config {
	return &Config{
		keyDelim: ".",
		format:   format,
		viper:    viper.New(),
	}
}

func formatFromExtension(configFile string) string {
	return strings.TrimPrefix(filepath.Ext(configFile), ".")
}

// LoadFile loads configuration from file
// Leave format blank for auto-detection from the file extension
func LoadFile(configFile, format string) (*Config, error) {
	if format == "" {
		format = formatFromExtension(configFile)
	}

	c := newConfig(format)
	c.configFile = configFile

	readAndAdd := func(configFile string, replace bool) error {
		clog.Debugf("loading: %s", configFile)
		file, err := os.Open(configFile)
		if err != nil {
			return fmt.Errorf("cannot open configuration file for reading: %w", err)
		}
		defer file.Close()

		return c.addTemplate(file, configFile, replace)
	}

	// Load config file
	err := readAndAdd(configFile, true)
	if err != nil {
		return c, err
	}

	// Load includes (if any).
	var includes []string
	if includes, err = filesearch.FindConfigurationIncludes(configFile, c.getIncludes()); err == nil {
		for _, include := range includes {
			format := formatFromExtension(include)

			switch {
			case format == FormatHCL && c.format != FormatHCL:
				err = fmt.Errorf("hcl format (%s) cannot be used in includes from %s: %s", include, c.format, c.configFile)
			case c.format == FormatHCL && format != FormatHCL:
				err = fmt.Errorf("%s is in hcl format, includes must use the same format: cannot load %s", c.configFile, include)
			default:
				err = readAndAdd(include, false)
				if err == nil {
					c.includeFiles = append(c.includeFiles, include)
				}
			}

			if err != nil {
				break
			}
		}
	}
	if err == nil && c.includeFiles != nil {
		err = c.loadTemplates()
	}

	return c, err
}

// Load configuration from reader
// This should only be used for unit tests
func Load(input io.Reader, format string) (*Config, error) {
	c := newConfig(format)
	err := c.addTemplate(input, c.configFile, true)
	if err != nil {
		return c, err
	}
	return c, nil
}

// getIncludes returns a list of configuration files to include in the current configuration
func (c *Config) getIncludes() []string {
	var files []string

	if c.IsSet(constants.SectionConfigurationIncludes) {
		includes := make([]string, 0)

		if err := c.unmarshalKey(constants.SectionConfigurationIncludes, &includes); err == nil {
			files = append(files, includes...)
		} else {
			clog.Errorf("Failed parsing includes definition: %v", err)
		}
	}

	return files
}

func (c *Config) templateName(name string) string {
	return "__config:" + name // prefixing name to avoid clash with named template defines
}

func (c *Config) addTemplate(input io.Reader, name string, replace bool) error {
	inputString := &strings.Builder{}
	_, err := io.Copy(inputString, input)
	if err != nil {
		return err
	}

	var source *template.Template
	if c.sourceTemplates == nil || replace {
		source = template.New(c.templateName(name))
		c.sourceTemplates = source
	} else {
		source = c.sourceTemplates.New(c.templateName(name))
	}

	_, err = source.Parse(inputString.String())
	if err != nil {
		return fmt.Errorf("cannot compile %w", err)
	}

	if replace {
		err = c.loadTemplates()
	}
	return err
}

// load configuration from an io.Reader
func (c *Config) load(input io.Reader, format string, replace bool) error {
	// For compatibility with the previous versions, a .conf file is TOML format
	if format == "conf" {
		format = "toml"
	}
	c.viper.SetConfigType(format)

	previousVersion := c.version
	c.version = VersionUnknown
	var err error
	if replace {
		err = c.viper.ReadConfig(input)
	} else {
		err = c.viper.MergeConfig(input)
	}

	if err != nil {
		return fmt.Errorf("cannot parse %s configuration: %w", format, err)
	}

	if previousVersion != c.GetVersion() && previousVersion > VersionUnknown {
		return errors.New("cannot include different versions of the configuration file, all files must use the same version")
	}
	return nil
}

func (c *Config) loadTemplates() error {
	return c.reloadTemplates(newTemplateData(c.configFile, "default", ""))
}

func (c *Config) reloadTemplates(data TemplateData) error {
	if c.sourceTemplates == nil {
		return errors.New("no available template to execute, please load it first")
	}

	buffer := &bytes.Buffer{}
	executeTemplate := func(name, format string, replace bool) error {
		buffer.Reset()
		err := c.sourceTemplates.ExecuteTemplate(buffer, c.templateName(name), data)
		if err != nil {
			return fmt.Errorf("cannot execute %w", err)
		}

		traceConfig(data.Profile.Name, name, replace, buffer.String())
		return c.load(buffer, format, replace)
	}

	// Load main config file
	var err error
	err = executeTemplate(c.configFile, c.format, true)

	// Load includes
	if err == nil && c.includeFiles != nil {
		for _, file := range c.includeFiles {
			err = executeTemplate(file, formatFromExtension(file), false)
			if err != nil {
				break
			}
		}
	}

	return err
}

// DisplayConfigurationIssues logs issues in the configuration for all profiles previously returned by GetProfile
func (c *Config) DisplayConfigurationIssues() {
	if len(c.issues.changedPaths) > 0 {
		var msg []string
		for path, resolved := range c.issues.changedPaths {
			msg = append(msg, fmt.Sprintf(`> %s changes to "%s"`, path, strings.Join(resolved, `", "`)))
		}
		sort.Strings(msg)
		msg = append([]string{
			"the configuration contains relative 'path' items which may lead to unstable results in restic " +
				"commands that select snapshots. Consider using absolute paths in 'path' (and 'source') or use " +
				"'tag' instead of 'path' (path = false) to select snapshots for restic commands. Affected paths:",
		}, msg...)
		clog.Info(strings.Join(msg, fmt.Sprintln()))
	}

	// Reset issues
	c.issues.changedPaths = nil
}

// IsSet checks if the key contains a value. Keys and subkeys can be separated by a "."
func (c *Config) IsSet(key string) bool {
	if strings.Contains(key, ".") && c.format == FormatHCL {
		clog.Error("HCL format is not supported in version 2, please use version 1 or another file format")
		return false
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
	return c.IsSet(c.getProfilePath(profileKey))
}

// GetProfileNames returns all profile names defined in the configuration
func (c *Config) GetProfileNames() []string {
	profiles := make([]string, 0)
	viperScope := c.viper
	if c.GetVersion() >= Version02 {
		// move to the profiles subsection
		viperScope = viperScope.Sub(constants.SectionConfigurationProfiles)
		if viperScope == nil {
			// there's no such subsection, so return the empty map
			return profiles
		}
	}
	allSettings := viperScope.AllSettings()
	for sectionKey := range allSettings {
		if c.GetVersion() == Version01 &&
			(sectionKey == constants.SectionConfigurationGlobal ||
				sectionKey == constants.SectionConfigurationGroups ||
				sectionKey == constants.SectionConfigurationIncludes) {
			continue
		}
		profiles = append(profiles, sectionKey)
	}
	return profiles
}

// GetProfiles returns a map of all available profiles with their configuration
func (c *Config) GetProfiles() map[string]*Profile {
	profiles := make(map[string]*Profile)
	for _, profileName := range c.GetProfileNames() {
		profile, err := c.GetProfile(profileName)
		if err != nil {
			clog.Error(err)
			continue
		}
		profiles[profileName] = profile
	}
	return profiles
}

// GetVersion returns the version of the configuration file.
// Default is Version01 if not specified or invalid
func (c *Config) GetVersion() Version {
	if c.version > VersionUnknown {
		return c.version
	}
	c.version = ParseVersion(c.viper.GetString(constants.ParameterVersion))
	return c.version
}

// GetGlobalSection returns the global configuration
func (c *Config) GetGlobalSection() (*Global, error) {
	global := NewGlobal()
	err := c.unmarshalKey(constants.SectionConfigurationGlobal, global)
	if err != nil {
		return nil, err
	}

	// All files in the configuration are relative to the configuration file,
	// NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	if rootPath != "." {
		rootPathMessage.Do(func() {
			clog.Debugf("files in configuration are relative to %q", rootPath)
		})
	}
	global.SetRootPath(rootPath)

	return global, nil
}

// HasProfileGroup returns true if the group of profiles exists in the configuration
func (c *Config) HasProfileGroup(groupKey string) bool {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return false
	}
	if err := c.loadGroups(); err != nil {
		return false
	}
	_, ok := c.groups[groupKey]
	return ok
}

// GetProfileGroup returns the list of profiles in a group
func (c *Config) GetProfileGroup(groupKey string) (*Group, error) {
	if err := c.loadGroups(); err != nil {
		return nil, err
	}

	group, ok := c.groups[groupKey]
	if !ok {
		return nil, fmt.Errorf("group '%s' not found", groupKey)
	}
	return &group, nil
}

// GetProfileGroups returns all groups from the configuration
//
// If the groups section does not exist, it returns an empty map
func (c *Config) GetProfileGroups() map[string]Group {
	if err := c.loadGroups(); err != nil {
		return nil
	}
	return c.groups
}

func (c *Config) loadGroups() error {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		c.groups = map[string]Group{}
		return nil
	}
	// load groups only once
	if c.groups == nil {
		if c.GetVersion() == Version01 {
			c.groups = map[string]Group{}
			groups := map[string][]string{}
			err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
			if err != nil {
				return err
			}
			// fits previous version into new structure
			for groupName, group := range groups {
				c.groups[groupName] = Group{
					Profiles: group,
				}
			}
			return nil
		}
		// Version 2 onwards
		groups := map[string]Group{}
		err := c.unmarshalKey(constants.SectionConfigurationGroups, &groups)
		if err != nil {
			return err
		}
		c.groups = groups
	}
	return nil
}

// GetProfile in configuration. If the profile is not found, it returns errNotFound
func (c *Config) GetProfile(profileKey string) (*Profile, error) {
	if c.sourceTemplates != nil {
		err := c.reloadTemplates(newTemplateData(c.configFile, profileKey, ""))
		if err != nil {
			return nil, err
		}
	}
	profile, err := c.getProfile(profileKey)
	if err != nil {
		return nil, err
	}
	// profile shouldn't be nil with no error, but better safe than sorry
	if profile == nil {
		return nil, errors.New("unexpected nil profile")
	}

	// Resolve config dependencies
	profile.ResolveConfiguration()

	// All files in the configuration are relative to the configuration file,
	// NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	profile.SetRootPath(rootPath)

	return profile, nil
}

// getProfile from configuration. If the profile is not found, it returns errNotFound
func (c *Config) getProfile(profileKey string) (*Profile, error) {
	var err error
	var profile *Profile

	if !c.IsSet(c.getProfilePath(profileKey)) {
		// profile key not found
		return nil, ErrNotFound
	}

	profile = NewProfile(c, profileKey)
	err = c.unmarshalKey(c.getProfilePath(profileKey), profile)
	if err != nil {
		return nil, err
	}

	if profile.Inherit != "" {
		inherit := profile.Inherit
		// Load inherited profile
		profile, err = c.getProfile(inherit)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, fmt.Errorf("error in profile '%s': parent profile '%s' not found", profileKey, inherit)
			}
			return nil, err
		}
		// It doesn't make sense to inherit the Description field
		profile.Description = ""
		// Reload this profile onto the inherited one
		err = c.unmarshalKey(c.getProfilePath(profileKey), profile)
		if err != nil {
			return nil, err
		}
		// make sure it has the right name
		profile.Name = profileKey
	}

	// Hide confidential values (keys, passwords) from the public representation
	ProcessConfidentialValues(profile)

	return profile, nil
}

// getProfilePath returns the key prefixed with "profiles" if the configuration file version is >= 2
func (c *Config) getProfilePath(key string) string {
	if c.GetVersion() == Version01 {
		return key
	}
	return constants.SectionConfigurationProfiles + "." + key
}

// GetSchedules loads all schedules from the configuration.
// !!! Nothing is using this method yet !!!
func (c *Config) GetSchedules() ([]*ScheduleConfig, error) {
	if c.GetVersion() == Version01 {
		return c.getSchedulesV1()
	}
	return nil, nil
}

// getSchedulesV1 loads schedules from profiles
func (c *Config) getSchedulesV1() ([]*ScheduleConfig, error) {
	profiles := c.GetProfileNames()
	if len(profiles) == 0 {
		return nil, nil
	}
	schedules := []*ScheduleConfig{}
	for _, profileName := range profiles {
		profile, err := c.GetProfile(profileName)
		if err != nil {
			return nil, fmt.Errorf("cannot load profile %q: %w", profileName, err)
		}
		profileSchedules := profile.Schedules()
		schedules = append(schedules, profileSchedules...)
	}
	return schedules, nil
}

// GetScheduleSections returns a list of schedules
func (c *Config) GetScheduleSections() (map[string]Schedule, error) {
	schedules := map[string]Schedule{}
	if c.GetVersion() < Version02 {
		return nil, errors.New("expected configuration >= version 2")
	}
	// move to the schedules subsection
	viperScope := c.viper.Sub(constants.SectionConfigurationSchedules)
	if viperScope == nil {
		// there's no such subsection, so return the empty map
		return schedules, nil
	}

	allSettings := viperScope.AllSettings()
	for sectionKey := range allSettings {
		schedule, err := c.getSchedule(sectionKey)
		if err != nil {
			return schedules, err
		}
		schedules[sectionKey] = schedule
	}
	return schedules, nil
}

func (c *Config) getSchedule(key string) (Schedule, error) {
	schedule := Schedule{}
	err := c.unmarshalKey(constants.SectionConfigurationSchedules+"."+key, &schedule)
	if err != nil {
		return schedule, err
	}
	return schedule, nil
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
		return data, nil
	}
}

// traceConfig sends a log of level trace to show the resulting configuration after resolving the template
func traceConfig(profileName, name string, replace bool, config string) {
	lines := strings.Split(config, "\n")
	output := ""
	for i := 0; i < len(lines); i++ {
		output += fmt.Sprintf("%3d: %s\n", i+1, lines[i])
	}
	clog.Tracef("Resulting configuration for profile '%s' ('%s' / replace=%v):\n"+
		"====================\n"+
		"%s"+
		"====================\n", profileName, name, replace, output)
}
