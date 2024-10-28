package config

import (
	"reflect"
	"strings"
)

type UpgradeSchedule struct {
}

// Upgrade all legacy schedule parameters into a schedule subsection
func (u *UpgradeSchedule) Upgrade(key string, config *Config) error {
	if !config.viper.IsSet(key) {
		return nil
	}
	schedulePrefix := "schedule-"
	scheduleFields := make([]string, 0, 10)
	scheduleParameters := reflect.TypeOf(ScheduleBaseSection{})
	for i := 0; i < scheduleParameters.NumField(); i++ {
		field := scheduleParameters.Field(i)
		tag := field.Tag.Get("mapstructure")
		if strings.HasPrefix(tag, schedulePrefix) {
			scheduleFields = append(scheduleFields, tag)
		}
	}

	commands := NewProfile(config, "").SchedulableCommands()
	for _, command := range commands {
		if config.viper.IsSet(config.flatKey(key, command, "schedule")) && !config.viper.IsSet(config.flatKey(key, command, "schedule", "at")) {
			if schedule := config.viper.GetString(config.flatKey(key, command, "schedule")); len(schedule) > 0 {
				config.viper.Set(config.flatKey(key, command, "schedule", "at"), schedule)

			} else if schedule := config.viper.GetStringSlice(config.flatKey(key, command, "schedule")); len(schedule) > 0 {
				config.viper.Set(config.flatKey(key, command, "schedule", "at"), schedule)
			}
		}
		for _, scheduleField := range scheduleFields {
			strippedField := strings.TrimPrefix(scheduleField, schedulePrefix)
			if config.viper.IsSet(config.flatKey(key, command, scheduleField)) && !config.viper.IsSet(config.flatKey(key, command, "schedule", strippedField)) {
				value := config.viper.Get(config.flatKey(key, command, scheduleField))
				config.viper.Set(config.flatKey(key, command, "schedule", strippedField), value)
				config.viper.Set(config.flatKey(key, command, scheduleField), nil)
			}
		}
	}
	return nil
}
