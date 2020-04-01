package config

import (
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/spf13/viper"
)

type Profile struct {
	Name        string
	Quiet       bool                   `mapstructure:"quiet"`
	Verbose     bool                   `mapstructure:"verbose"`
	Repository  string                 `mapstructure:"repository"`
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
