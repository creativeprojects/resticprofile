package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpgradeSchedules(t *testing.T) {
	raw := `
version: "2"
profiles:
  profile1:
    backup:
      schedule: daily
      schedule-permission: "user"
  profile2:
    backup:
      schedule:
        at: daily
        permission: "user"
  profile3:
    backup:
      schedule: [daily, weekly]
  profile4:
    backup:
      schedule:
        at: [daily, weekly]
`
	config, err := Load(bytes.NewBufferString(raw), FormatYAML)
	require.NoError(t, err)
	assert.NotNil(t, config)

	for _, profileName := range []string{"profile1", "profile2"} {
		t.Run(profileName, func(t *testing.T) {
			profile := config.flatKey("profiles", profileName)
			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule")))

			new(UpgradeSchedule).Upgrade(profile, config)

			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule")))
			assert.Equal(t,
				map[string]interface{}{"at": "daily", "permission": "user"},
				config.viper.Get(config.flatKey(profile, "backup", "schedule")))
			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule", "at")))
			assert.Equal(t, "daily", config.viper.Get(config.flatKey(profile, "backup", "schedule", "at")))
		})
	}

	for _, profileName := range []string{"profile3", "profile4"} {
		t.Run(profileName, func(t *testing.T) {
			profile := config.flatKey("profiles", profileName)
			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule")))

			new(UpgradeSchedule).Upgrade(profile, config)

			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule")))
			assert.True(t, config.viper.IsSet(config.flatKey(profile, "backup", "schedule", "at")))
			assert.EqualValues(t, []string{"daily", "weekly"}, config.viper.GetStringSlice(config.flatKey(profile, "backup", "schedule", "at")))
		})
	}
}
