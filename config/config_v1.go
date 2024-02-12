package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// This is where things are getting hairy:
//
// Most configuration file formats allow only one declaration per section
// This is not the case for HCL where you can declare a bloc multiple times:
//
//	"global" {
//	  key1 = "value"
//	}
//
//	"global" {
//	  key2 = "value"
//	}
//
// For that matter, viper creates a slice of maps instead of a map for the other configuration file formats
// This configOptionV1HCL deals with the slice to merge it into a single map
var (
	configOptionV1 = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		maybe.BoolDecoder(),
		confidentialValueDecoder(),
	))

	configOptionV1HCL = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		maybe.BoolDecoder(),
		confidentialValueDecoder(),
		sliceOfMapsToMapHookFunc(),
	))
)

// getProfileNamesV1 returns all profile names defined in the configuration version 1
func (c *Config) getProfileNamesV1() (names []string) {
	c.requireVersion(Version01)

	names = make([]string, 0)
	for sectionKey := range c.viper.AllSettings() {
		if sectionKey == constants.SectionConfigurationGlobal ||
			sectionKey == constants.SectionConfigurationGroups ||
			sectionKey == constants.SectionConfigurationIncludes ||
			sectionKey == constants.ParameterVersion ||
			sectionKey == constants.JSONSchema {
			continue
		}
		names = append(names, sectionKey)
	}
	return
}

func (c *Config) loadGroupsV1() (err error) {
	c.requireVersion(Version01)

	if c.groups == nil {
		c.groups = map[string]Group{}

		if c.IsSet(constants.SectionConfigurationGroups) {
			groups := map[string][]string{}
			if err = c.unmarshalKey(constants.SectionConfigurationGroups, &groups); err == nil {
				// fits previous version into new structure
				for groupName, group := range groups {
					c.groups[groupName] = Group{
						Description:     "",
						Profiles:        group,
						ContinueOnError: maybe.Bool{},
					}
				}
			}
		}
	}
	return err
}

// getProfileV1 from version 1 configuration. If the profile is not found, it returns errNotFound
func (c *Config) getProfileV1(profileKey string) (profile *Profile, err error) {
	c.requireVersion(Version01)

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

	return profile, nil
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

// unmarshalConfigV1 returns the viper.DecoderConfigOption to use for V1 configuration files
func (c *Config) unmarshalConfigV1() viper.DecoderConfigOption {
	c.requireVersion(Version01)

	if c.format == "hcl" {
		return configOptionV1HCL
	} else {
		return configOptionV1
	}
}

// sliceOfMapsToMapHookFunc merges a slice of maps to a map
func sliceOfMapsToMapHookFunc() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Slice && from.Elem().Kind() == reflect.Map {
			// unpack single slice always (needed for nested maps like OtherFlags)
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
			// flatten slice of maps into one map
			if to.Kind() == reflect.Struct || to.Kind() == reflect.Map {
				convert := make(map[string]interface{})
				for _, mapItem := range source {
					for key, value := range mapItem {
						convert[key] = value
					}
				}
				return convert, nil
			}
		}
		return data, nil
	}
}
