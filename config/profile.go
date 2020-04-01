package config

import (
	"reflect"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

type Profile struct {
	Name        string
	Quiet       bool                   `mapstructure:"quiet" argument:"quiet"`
	Verbose     bool                   `mapstructure:"verbose" argument:"verbose"`
	Repository  string                 `mapstructure:"repository" argument:"repo"`
	Initialize  bool                   `mapstructure:"initialize"`
	Inherit     string                 `mapstructure:"inherit"`
	Lock        bool                   `mapstructure:"lock"`
	Environment map[string]string      `mapstructure:"env"`
	Backup      *BackupSection         `mapstructure:"backup"`
	Retention   *RetentionSection      `mapstructure:"retention"`
	Snapshots   map[string]interface{} `mapstructure:"snapshots"`
	Forget      map[string]interface{} `mapstructure:"forget"`
	Check       map[string]interface{} `mapstructure:"check"`
	Mount       map[string]interface{} `mapstructure:"mount"`
	OtherFlags  map[string]interface{} `mapstructure:",remain"`
}

type BackupSection struct {
	CheckBefore bool                   `mapstructure:"check-before"`
	CheckAfter  bool                   `mapstructure:"check-after"`
	RunBefore   []string               `mapstructure:"run-before"`
	RunAfter    []string               `mapstructure:"run-after"`
	UseStdin    bool                   `mapstructure:"stdin"`
	Source      []string               `mapstructure:"source"`
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

func (p *Profile) GetCommonFlags() map[string]string {
	flags := make(map[string]string, 0)
	pt := reflect.TypeOf(*p)

	// NumField() will panic if pt is not a struct
	if pt.Kind() != reflect.Struct {
		return flags
	}
	for i := 0; i < pt.NumField(); i++ {
		field := pt.Field(i)
		if argument, ok := field.Tag.Lookup("argument"); ok {
			if argument != "" {
				convert, ok := stringify("set") // <-- find value of field
				if ok {
					flags[argument] = convert
				}
			}
		}
	}
	if p.OtherFlags == nil || len(p.OtherFlags) == 0 {
		return flags
	}

	// Add other flags
	for name, value := range p.OtherFlags {
		if convert, ok := stringify(value); ok {
			flags[name] = convert
		}
	}
	return flags
}

func (p *Profile) GetCommandFlags(command string) []string {
	flags := make([]string, 0)
	return flags
}

func (p *Profile) GetRetentionFlags() []string {
	flags := make([]string, 0)
	return flags
}

func (p *Profile) GetBackupSource() []string {
	sourcePaths := make([]string, 0)
	return sourcePaths
}

func stringify(value interface{}) (string, bool) {
	if convert, ok := value.(string); ok {
		if convert != "" {
			return convert, true
		}
		return "", false
	}
	if convert, ok := value.(bool); ok && convert {
		return "", true
	}
	return "", false
}
