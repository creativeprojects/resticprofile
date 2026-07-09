package config

import (
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasTriggers(t *testing.T) {
	t.Run("nil is false", func(t *testing.T) {
		var sc *ScheduleConfig
		assert.False(t, sc.HasTriggers())
	})

	t.Run("calendar schedule has triggers", func(t *testing.T) {
		sc := &ScheduleConfig{Schedules: []string{"daily"}}
		assert.True(t, sc.HasSchedules())
		assert.True(t, sc.HasTriggers())
	})

	t.Run("after-login only has triggers but no schedules", func(t *testing.T) {
		sc := &ScheduleConfig{}
		sc.AfterLogin = maybe.True()
		assert.False(t, sc.HasSchedules())
		assert.True(t, sc.HasTriggers())
	})

	t.Run("empty has no triggers", func(t *testing.T) {
		sc := &ScheduleConfig{}
		assert.False(t, sc.HasTriggers())
	})
}

func TestAfterLoginScheduleNotDropped(t *testing.T) {
	profile := func(t *testing.T, config string) *Profile {
		t.Helper()
		if !strings.Contains(config, "[default") {
			config += "\n[default]"
		}
		profile, err := getResolvedProfile("toml", config, "default")
		require.NoError(t, err)
		require.NotNil(t, profile)
		return profile
	}

	t.Run("struct form with after-login only", func(t *testing.T) {
		p := profile(t, `
			[default.backup]
			schedule-permission = "user_logged_on"
			[default.backup.schedule]
			after-login = true
		`)
		schedules := p.Schedules()
		require.Contains(t, schedules, "backup")
		sc := schedules["backup"]
		assert.True(t, sc.AfterLogin.IsTrue())
		assert.False(t, sc.HasSchedules())
		assert.True(t, sc.HasTriggers())
		assert.Equal(t, constants.SchedulePermissionUserLoggedOn, sc.Permission)
	})

	t.Run("flat schedule-after-login override coexists with at times", func(t *testing.T) {
		p := profile(t, `
			[default.backup]
			schedule = "daily"
			schedule-after-login = true
		`)
		schedules := p.Schedules()
		require.Contains(t, schedules, "backup")
		sc := schedules["backup"]
		assert.True(t, sc.AfterLogin.IsTrue())
		assert.Equal(t, []string{"daily"}, sc.Schedules)
		assert.True(t, sc.HasTriggers())
	})
}
