package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/filesearch"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config wraps up a viper configuration object
type Config struct {
	keyDelim        string
	format          string
	configFile      string
	includeFiles    []string
	envFiles        []string
	lastProfileKey  string
	viper           *viper.Viper
	mixinUses       []map[string][]*mixinUse
	mixins          map[string]*mixin
	sourceTemplates *template.Template
	version         Version
	issues          struct {
		changedPaths  map[string][]string // 'path' items that had been changed to absolute paths
		failedSection map[string]error    // profile sections that failed to get parsed or resolved
	}
	cached struct {
		groups map[string]*Group
		global *Global
	}
}

var (
	configOption = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		maybe.BoolDecoder(),
		maybe.DurationDecoder(),
		confidentialValueDecoder(),
	))

	rootPathMessage = sync.Once{}
)

// newConfig instantiate a new Config object
func newConfig(format string) *Config {
	keyDelimiter := "\\"
	return &Config{
		keyDelim: keyDelimiter,
		format:   format,
		viper:    viper.NewWithOptions(viper.KeyDelimiter(keyDelimiter)),
	}
}

func formatFromExtension(configFile string) string {
	return strings.TrimPrefix(filepath.Ext(configFile), ".")
}

// LoadFile loads configuration from file
// Leave format blank for auto-detection from the file extension
func LoadFile(configFile, format string) (config *Config, err error) {
	if format == "" {
		format = formatFromExtension(configFile)
	}

	config = newConfig(format)
	config.configFile = configFile

	readAndAdd := func(configFile string, replace bool) error {
		clog.Debugf("loading: %s", configFile)
		file, fileErr := os.Open(configFile)
		if fileErr != nil {
			return fmt.Errorf("cannot open configuration file for reading: %w", fileErr)
		}
		defer file.Close()

		return config.addTemplate(file, configFile, replace)
	}

	// Load config file
	err = readAndAdd(configFile, true)
	if err != nil {
		return
	}

	// Load includes (if any).
	var includes []string
	if includes, err = filesearch.NewFinder().FindConfigurationIncludes(configFile, config.getIncludes()); err == nil {
		for _, include := range includes {
			format := formatFromExtension(include)

			switch {
			case format == FormatHCL && config.format != FormatHCL:
				err = fmt.Errorf("hcl format (%s) cannot be used in includes from %s: %s", include, config.format, config.configFile)
			case config.format == FormatHCL && format != FormatHCL:
				err = fmt.Errorf("%s is in hcl format, includes must use the same format: cannot load %s", config.configFile, include)
			default:
				err = readAndAdd(include, false)
				if err == nil {
					config.includeFiles = append(config.includeFiles, include)
				}
			}

			if err != nil {
				break
			}
		}
	}
	if err == nil && config.includeFiles != nil {
		err = config.loadTemplates()
	}

	return
}

// Load configuration from reader
// This should only be used for unit tests
func Load(input io.Reader, format string, options ...func(cfg *Config)) (config *Config, err error) {
	config = newConfig(format)
	for _, option := range options {
		option(config)
	}
	err = config.addTemplate(input, config.configFile, true)
	return
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
	if rs, ok := input.(io.ReadSeeker); ok {
		input = util.NewUTF8Reader(rs)
	}

	inputString := &strings.Builder{}
	_, err := io.Copy(inputString, input)
	if err != nil {
		return err
	}

	var source *template.Template
	if c.sourceTemplates == nil || replace {
		envFile := templates.EnvFileFunc(func() (string, func(string)) { return c.lastProfileKey, c.addEnvFile })
		source = templates.New(c.templateName(name), envFile)
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
func (c *Config) load(input io.Reader, format string, replace bool) (err error) {
	if format == "conf" { // A .conf file is TOML format
		format = "toml"
	}

	previousVersion := c.version
	c.version = VersionUnknown

	var vp *viper.Viper
	if replace {
		c.mixinUses = nil
		vp = c.viper
	} else {
		vp = newConfig(format).viper
	}

	vp.SetConfigType(format)
	err = vp.ReadConfig(input)

	if err == nil && vp != c.viper {
		err = c.viper.MergeConfigMap(vp.AllSettings())
	}

	if err == nil && c.GetVersion() >= Version02 {
		var allUses map[string][]*mixinUse
		if allUses, err = collectAllMixinUses(vp, c.keyDelim); err == nil && len(allUses) > 0 {
			c.mixinUses = append(c.mixinUses, allUses)
		}
	}

	if err != nil {
		return fmt.Errorf("cannot parse %s configuration: %w", format, err)
	}

	if previousVersion != c.GetVersion() && previousVersion > VersionUnknown {
		err = errors.New("cannot include different versions of the configuration file, all files must use the same version")
	}
	return
}

func (c *Config) applyNonProfileMixins() error {
	return c.applyMatchingMixinsOnce(func(useKey string) bool {
		return !strings.HasPrefix(useKey, constants.SectionConfigurationProfiles)
	})
}

func (c *Config) applyMixinsToProfile(profileName string) error {
	prefix := c.getProfilePath(profileName)
	return c.applyMatchingMixinsOnce(func(useKey string) bool {
		return strings.HasPrefix(useKey, prefix)
	})
}

func (c *Config) applyMatchingMixinsOnce(matcher func(useKey string) bool) error {
	var matchingUses []map[string][]*mixinUse

	for _, allUses := range c.mixinUses {
		usesToApply := make(map[string][]*mixinUse)
		matchingUses = append(matchingUses, usesToApply)

		for useKey, uses := range allUses {
			if matcher(useKey) {
				usesToApply[useKey] = uses
				delete(allUses, useKey) // remove mixinUse to avoid double apply
			}
		}
	}

	if len(matchingUses) > 0 {
		return c.applyMixins(matchingUses)
	}
	return nil
}

func (c *Config) applyMixins(allUsesToApply []map[string][]*mixinUse) (err error) {
	c.requireMinVersion(Version02)

	if allUsesToApply == nil {
		allUsesToApply = c.mixinUses
	}

	for _, uses := range allUsesToApply {
		if err = applyMixins(c.viper, c.keyDelim, uses, c.mixins); err != nil {
			break
		}
	}
	return
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

		traceConfig(data.Profile.Name, name, replace, buffer)
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

	// Load mixins and apply outside of profiles
	if err == nil && c.GetVersion() >= Version02 {
		c.mixins = parseMixins(c.viper)
		err = c.applyNonProfileMixins()
	}

	// clear cached items
	c.cached.groups = nil
	c.cached.global = nil

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
			"the configuration contains relative \"path\" items which may lead to unstable results in restic " +
				"commands that select snapshots. Consider using absolute paths in \"path\" (and \"source\"), " +
				"set \"base-dir\" or \"source-base\" in the profile or use \"tag\" instead of \"path\" (path = false) to select " +
				"snapshots for restic commands.",
			"Affected paths are:",
		}, msg...)
		clog.Info(strings.Join(msg, fmt.Sprintln()))
	}

	if len(c.issues.failedSection) > 0 {
		names := slices.Sorted(maps.Keys(c.issues.failedSection))
		for _, name := range names {
			clog.Errorf("Failed parsing profile section %q: %s", name, c.issues.failedSection[name].Error())
		}
	}

	allNames := make(map[string]bool)
	duplicates := make([]string, 0)
	for _, name := range append(c.GetProfileNames(), c.GetGroupNames()...) {
		if allNames[name] {
			duplicates = append(duplicates, name)
		} else {
			allNames[name] = true
		}
	}
	if len(duplicates) > 0 {
		clog.Errorf("Duplicate profile and group names found: %s. This could lead to unpredictable behaviour.", strings.Join(duplicates, ", "))
	}

	// Reset issues
	c.ClearConfigurationIssues()
}

// ClearConfigurationIssues removes all configuration issues
func (c *Config) ClearConfigurationIssues() {
	c.issues.changedPaths = nil
	c.issues.failedSection = nil
}

func (c *Config) reportChangedPath(resolvedPath, path, origin string) {
	if c.issues.changedPaths == nil {
		c.issues.changedPaths = make(map[string][]string)
	}
	key := fmt.Sprintf(`%s "%s"`, origin, path)
	c.issues.changedPaths[key] = append(c.issues.changedPaths[key], resolvedPath)
}

func (c *Config) reportFailedSection(name string, err error) {
	if c.issues.failedSection == nil {
		c.issues.failedSection = make(map[string]error)
	}
	c.issues.failedSection[name] = err
}

func (c *Config) flatKey(key ...string) (fk string) {
	if len(key) > 0 {
		fk = key[0]
		if len(key) > 1 {
			fk = strings.Join(key, c.keyDelim)
		}
	}
	return
}

// IsSet checks if the key contains a value.
// Key and subkey can be queried with IsSet(key, subkey) or by separating them with keyDelim.
func (c *Config) IsSet(key ...string) bool {
	flatKey := c.flatKey(key...)

	if strings.Contains(flatKey, c.keyDelim) && c.format == FormatHCL {
		clog.Error("HCL format is not supported in version 2, please use version 1 or another file format")
		return false
	}

	return c.viper.IsSet(flatKey)
}

// GetConfigFile returns the config file used
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// Get the value from the key
func (c *Config) Get(key ...string) interface{} {
	return c.viper.Get(c.flatKey(key...))
}

// HasProfile returns true if the profile exists in the configuration
func (c *Config) HasProfile(profileKey string) bool {
	return c.IsSet(c.getProfilePath(profileKey))
}

// GetProfileNames returns all profile names defined in the configuration
func (c *Config) GetProfileNames() (names []string) {
	if c.GetVersion() <= Version01 {
		return c.getProfileNamesV1()
	}

	names = make([]string, 0)
	if profiles := c.viper.Sub(constants.SectionConfigurationProfiles); profiles != nil {
		for sectionKey := range profiles.AllSettings() {
			names = append(names, sectionKey)
		}
	}
	return
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

func (c *Config) requireVersion(version Version) {
	if c.GetVersion() != version {
		panic(fmt.Sprintf("invalid api usage: expected config version %d, found %d", version, c.GetVersion()))
	}
}

func (c *Config) requireMinVersion(version Version) {
	if c.GetVersion() < version {
		panic(fmt.Sprintf("invalid api usage: expected min config version %d, found %d", version, c.GetVersion()))
	}
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
		rootPathMessage.Do(func() { clog.Debugf("files in configuration are relative to %q", rootPath) })
	}
	global.SetRootPath(rootPath)

	return global, nil
}

// mustGetGlobalSection returns a cached global configuration, panics if it can't be loaded (for internal use only)
func (c *Config) mustGetGlobalSection() *Global {
	if c.cached.global == nil {
		var err error
		c.cached.global, err = c.GetGlobalSection()
		if err != nil {
			panic(fmt.Errorf("MustGetGlobalSection: %w", err))
		}
	}
	return c.cached.global
}

// HasProfileGroup returns true if the group of profiles exists in the configuration
func (c *Config) HasProfileGroup(groupKey string) bool {
	if !c.IsSet(constants.SectionConfigurationGroups) {
		return false
	}
	if err := c.loadGroups(); err != nil {
		return false
	}
	_, ok := c.cached.groups[groupKey]
	return ok
}

// GetProfileGroup returns the list of profiles in a group
func (c *Config) GetProfileGroup(groupKey string) (*Group, error) {
	if err := c.loadGroups(); err != nil {
		return nil, err
	}

	group, ok := c.cached.groups[groupKey]
	if !ok {
		return nil, ErrNotFound
	}
	return group, nil
}

// GetProfileGroups returns all groups from the configuration
//
// If the groups section does not exist, it returns an empty map
func (c *Config) GetProfileGroups() map[string]*Group {
	if err := c.loadGroups(); err != nil {
		clog.Errorf("failed loading groups: %s", err.Error())
	}
	return maps.Clone(c.cached.groups)
}

func (c *Config) GetGroupNames() (names []string) {
	if c.GetVersion() <= Version01 {
		_ = c.loadGroupsV1()
		names = slices.Collect(maps.Keys(c.cached.groups))
	} else {
		if groups := c.viper.Sub(constants.SectionConfigurationGroups); groups != nil {
			for name := range groups.AllSettings() {
				names = append(names, name)
			}
		}
	}
	return
}

func (c *Config) loadGroups() (err error) {
	if c.GetVersion() <= Version01 {
		return c.loadGroupsV1()
	}

	if c.cached.groups == nil {
		c.cached.groups = make(map[string]*Group)
		for _, name := range c.GetGroupNames() {
			group := NewGroup(c, name)
			err = c.unmarshalKey(c.flatKey(constants.SectionConfigurationGroups, name), group)
			if err == nil {
				group.ResolveConfiguration()
				c.cached.groups[name] = group
			}
		}
	}
	return
}

func (c *Config) addEnvFile(filename string) {
	if !slices.Contains(c.envFiles, filename) {
		c.envFiles = append(c.envFiles, filename)
	}
}

// GetProfile in configuration. If the profile is not found, it returns errNotFound
func (c *Config) GetProfile(profileKey string) (profile *Profile, err error) {
	c.ClearConfigurationIssues()
	c.envFiles = nil
	c.lastProfileKey = profileKey

	if c.sourceTemplates != nil {
		err = c.reloadTemplates(newTemplateData(c.configFile, profileKey, ""))
		if err != nil {
			return
		}
	}

	profile, err = c.getProfile(profileKey)
	if err != nil {
		return
	}

	if profile == nil {
		// profile shouldn't be nil with no error, but better safe than sorry
		err = errors.New("unexpected nil profile")
		return
	}

	c.postProcessProfile(profile)
	return
}

// postProcessProfile applies additional post-processing steps before a profile can be used
func (c *Config) postProcessProfile(profile *Profile) {
	// Hide confidential values (keys, passwords) from the public representation
	ProcessConfidentialValues(profile)

	// Add additional envFiles
	profile.EnvironmentFiles = append(profile.EnvironmentFiles, c.envFiles...)

	// Resolve config dependencies
	profile.ResolveConfiguration()

	// All files in the configuration are relative to the configuration file,
	// NOT the folder where resticprofile is started
	// So we need to fix all relative files
	rootPath := filepath.Dir(c.GetConfigFile())
	profile.SetRootPath(rootPath)
}

func (c *Config) applyProfileInheritanceAndMixins(profileName string) (err error) {
	c.requireMinVersion(Version02)

	profilePath := c.getProfilePath(profileName)
	if !c.IsSet(profilePath) {
		err = ErrNotFound
		return
	}

	if inherit := c.viper.GetString(c.flatKey(profilePath, constants.SectionConfigurationInherit)); len(inherit) > 0 {

		inheritPath := c.getProfilePath(inherit)
		if !c.IsSet(inheritPath) {
			err = ErrNotFound
		} else {
			err = c.applyProfileInheritanceAndMixins(inherit) // recursive inheritance, the deepest first
		}

		if err == nil {
			// create merged profile for: parent > derived
			mergedProfile := viper.NewWithOptions(viper.KeyDelimiter(c.keyDelim))

			// init with parent (excluding some fields that must never be inherited)
			parent := c.viper.GetStringMap(inheritPath)
			delete(parent, constants.SectionConfigurationDescription)
			delete(parent, constants.SectionConfigurationMixinUse)
			delete(parent, constants.SectionConfigurationInherit)

			if err = mergedProfile.MergeConfigMap(parent); err == nil {
				// Merge derived onto parent (removing "inherit" instruction to ensure it is done only once)
				derived := c.viper.GetStringMap(profilePath)
				derived[constants.SectionConfigurationInherit] = ""
				revolveAppendToListKeys(mergedProfile, derived)

				err = mergedProfile.MergeConfigMap(derived)
			}
			if err != nil {
				return
			}

			// apply merged profile to config
			err = mergeConfigMap(c.viper, profilePath, c.keyDelim, mergedProfile.AllSettings())
		}

		if errors.Is(err, ErrNotFound) {
			err = fmt.Errorf("error in profile '%s': parent profile '%s' not found", profileName, inherit)
		}
	}

	// apply mixins
	if err == nil {
		err = c.applyMixinsToProfile(profileName)
	}
	return
}

// getProfile from configuration. If the profile is not found, it returns errNotFound
func (c *Config) getProfile(profileKey string) (profile *Profile, err error) {
	if c.GetVersion() <= Version01 {
		return c.getProfileV1(profileKey)
	}

	if err = c.applyProfileInheritanceAndMixins(profileKey); err != nil {
		return
	}

	profile = NewProfile(c, profileKey)
	err = c.unmarshalKey(c.getProfilePath(profileKey), profile)
	if err != nil {
		profile = nil
	}
	return
}

// getProfilePath returns the key prefixed with "profiles" if the configuration file version is >= 2
func (c *Config) getProfilePath(key string) string {
	if c.GetVersion() <= Version01 {
		return key
	}
	return c.flatKey(constants.SectionConfigurationProfiles, key)
}

// HasRemote returns true if the remote exists in the configuration
func (c *Config) HasRemote(remoteName string) bool {
	return c.IsSet(c.flatKey(constants.SectionConfigurationRemotes, remoteName))
}

func (c *Config) GetRemote(remoteName string) (*Remote, error) {
	// we don't need to check the file version: the remotes can be in a separate configuration file

	remote := NewRemote(c, remoteName)
	err := c.unmarshalKey(c.flatKey(constants.SectionConfigurationRemotes, remoteName), remote)

	rootPath := filepath.Dir(c.GetConfigFile())
	remote.SetRootPath(rootPath)
	return remote, err
}

// unmarshalConfig returns the decoder config options depending on the configuration version and format
func (c *Config) unmarshalConfig() viper.DecoderConfigOption {
	if c.GetVersion() == Version01 {
		return c.unmarshalConfigV1()
	} else {
		return configOption
	}
}

// unmarshalKey is a wrapper around viper.UnmarshalKey with the right decoder config options
func (c *Config) unmarshalKey(key string, rawVal interface{}) error {
	if c.GetVersion() >= Version02 && c.format == FormatHCL {
		return fmt.Errorf("HCL format is not supported in version %d, please use version 1 or another file format", c.GetVersion())
	}

	return c.viper.UnmarshalKey(key, rawVal, c.unmarshalConfig())
}

// newUnmarshaller returns a configured unmarshaler for output
func (c *Config) newUnmarshaller(output any) (*mapstructure.Decoder, error) {
	conf := &mapstructure.DecoderConfig{
		Result:           output,
		WeaklyTypedInput: true,
	}
	c.unmarshalConfig()(conf)

	return mapstructure.NewDecoder(conf)
}

// traceConfig sends a log of level trace to show the resulting configuration after resolving the template
func traceConfig(profileName, name string, replace bool, config *bytes.Buffer) {
	clog.Trace(func() string {
		output := &strings.Builder{}

		lines := bytes.Split(config.Bytes(), []byte("\n"))
		gutter := "%3d: "
		if len(lines) > 999 {
			gutter = "%4d: "
		}
		for i := 0; i < len(lines); i++ {
			output.WriteString(fmt.Sprintf(gutter, i+1))
			output.Write(lines[i])
			output.WriteString("\n")
		}

		return fmt.Sprintf(""+
			"Resulting configuration for profile '%s' ('%s' / replace=%v):\n"+
			"====================\n"+
			"%s"+
			"====================\n", profileName, name, replace, output)
	})
}
