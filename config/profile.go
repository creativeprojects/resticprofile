package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

type Profile struct {
	Name          string
	Quiet         bool                   `mapstructure:"quiet" argument:"quiet"`
	Verbose       bool                   `mapstructure:"verbose" argument:"verbose"`
	Repository    string                 `mapstructure:"repository" argument:"repo"`
	PasswordFile  string                 `mapstructure:"password-file" argument:"password-file"`
	CacheDir      string                 `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert        string                 `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert string                 `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize    bool                   `mapstructure:"initialize"`
	Inherit       string                 `mapstructure:"inherit"`
	Lock          string                 `mapstructure:"lock"`
	Environment   map[string]string      `mapstructure:"env"`
	Backup        *BackupSection         `mapstructure:"backup"`
	Retention     *RetentionSection      `mapstructure:"retention"`
	Snapshots     map[string]interface{} `mapstructure:"snapshots"`
	Forget        map[string]interface{} `mapstructure:"forget"`
	Check         map[string]interface{} `mapstructure:"check"`
	Mount         map[string]interface{} `mapstructure:"mount"`
	OtherFlags    map[string]interface{} `mapstructure:",remain"`
}

type BackupSection struct {
	CheckBefore bool                   `mapstructure:"check-before"`
	CheckAfter  bool                   `mapstructure:"check-after"`
	RunBefore   []string               `mapstructure:"run-before"`
	RunAfter    []string               `mapstructure:"run-after"`
	UseStdin    bool                   `mapstructure:"stdin" argument:"stdin"`
	Source      []string               `mapstructure:"source"`
	ExcludeFile []string               `mapstructure:"exclude-file" argument:"exclude-file"`
	FilesFrom   []string               `mapstructure:"files-from" argument:"files-from"`
	OtherFlags  map[string]interface{} `mapstructure:",remain"`
}

type RetentionSection struct {
	BeforeBackup bool                   `mapstructure:"before-backup"`
	AfterBackup  bool                   `mapstructure:"after-backup"`
	OtherFlags   map[string]interface{} `mapstructure:",remain"`
}

func NewProfile(name string) *Profile {
	return &Profile{
		Name: name,
	}
}

func LoadProfile(profileKey string) (*Profile, error) {
	var err error
	var profile *Profile

	if !viper.IsSet(profileKey) {
		return nil, nil
	}

	// Load parent profile first if it is inherited
	if viper.IsSet(profileKey + "." + constants.ParameterInherit) {
		inherit := viper.GetString(profileKey + "." + constants.ParameterInherit)
		if inherit != "" {
			profile, err = LoadProfile(inherit)
			if err != nil {
				return nil, err
			}
			if profile == nil {
				return nil, fmt.Errorf("Error in profile '%s': Parent profile '%s' not found", profileKey, inherit)
			}
			profile.Name = profileKey
		}
	}

	// If profile is not inherited, create a blank one
	if profile == nil {
		profile = NewProfile(profileKey)
	}

	err = viper.UnmarshalKey(profileKey, profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (p *Profile) SetRootPath(rootPath string) {

	p.Repository = fixPath(p.Repository, rootPath)
	p.PasswordFile = fixPath(p.PasswordFile, rootPath)
	p.CacheDir = fixPath(p.CacheDir, rootPath)
	p.CACert = fixPath(p.CACert, rootPath)
	p.TLSClientCert = fixPath(p.TLSClientCert, rootPath)

	if p.Backup != nil {
		if p.Backup.ExcludeFile != nil && len(p.Backup.ExcludeFile) > 0 {
			for i := 0; i < len(p.Backup.ExcludeFile); i++ {
				p.Backup.ExcludeFile[i] = fixPath(p.Backup.ExcludeFile[i], rootPath)
			}
		}

		if p.Backup.FilesFrom != nil && len(p.Backup.FilesFrom) > 0 {
			for i := 0; i < len(p.Backup.FilesFrom); i++ {
				p.Backup.FilesFrom[i] = fixPath(p.Backup.FilesFrom[i], rootPath)
			}
		}

		// Do we need to do source files? (it wasn't the case before v0.6.0)
		if p.Backup.Source != nil && len(p.Backup.Source) > 0 {
			for i := 0; i < len(p.Backup.Source); i++ {
				if filepath.IsAbs(p.Backup.Source[i]) {
					continue
				}
				p.Backup.Source[i] = fixPath(p.Backup.Source[i], rootPath)
			}
		}
	}
}

func (p *Profile) GetCommonFlags() map[string][]string {
	// Flags from the profile fields
	flags := convertStructToFlags(*p)

	flags = addOtherFlags(flags, p.OtherFlags)

	return flags
}

func (p *Profile) GetCommandFlags(command string) map[string][]string {
	flags := p.GetCommonFlags()

	switch command {
	case constants.CommandBackup:
		commandFlags := convertStructToFlags(*p.Backup)
		if commandFlags != nil && len(commandFlags) > 0 {
			flags = mergeFlags(flags, commandFlags)
		}
		flags = addOtherFlags(flags, p.Backup.OtherFlags)

	case constants.CommandSnapshots:
		flags = addOtherFlags(flags, p.Snapshots)

	case constants.CommandCheck:
		flags = addOtherFlags(flags, p.Check)
	}

	return flags
}

func (p *Profile) GetRetentionFlags() map[string][]string {
	flags := p.GetCommonFlags()
	flags = addOtherFlags(flags, p.Retention.OtherFlags)
	return flags
}

func (p *Profile) GetBackupSource() []string {
	return p.Backup.Source
}

func addOtherFlags(flags map[string][]string, otherFlags map[string]interface{}) map[string][]string {
	if otherFlags == nil || len(otherFlags) == 0 {
		return flags
	}

	// Add other flags
	for name, value := range otherFlags {
		if convert, ok := stringifyValueOf(value); ok {
			flags[name] = convert
		}
	}
	return flags
}

func mergeFlags(flags, newFlags map[string][]string) map[string][]string {
	if (flags == nil || len(flags) == 0) && newFlags != nil {
		return newFlags
	}
	if flags != nil && (newFlags == nil || len(newFlags) == 0) {
		return flags
	}
	for key, value := range newFlags {
		flags[key] = value
	}
	return flags
}

func fixPath(source, prefix string) string {
	if strings.Contains(source, "$") || strings.Contains(source, "%") {
		source = os.ExpandEnv(source)
	}
	if source == "" ||
		filepath.IsAbs(source) ||
		strings.HasPrefix(source, "~") ||
		strings.HasPrefix(source, "$") ||
		strings.HasPrefix(source, "%") {
		return source
	}
	return filepath.Join(prefix, source)
}
