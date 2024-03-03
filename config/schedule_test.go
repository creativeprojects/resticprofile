package config

import (
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchedule(t *testing.T) {
	profile := func(t *testing.T, config string) (*Profile, ScheduleConfigOrigin) {
		if !strings.Contains(config, "[default") {
			config += "\n[default]"
		}
		profile, err := getResolvedProfile("toml", config, "default")
		require.NoError(t, err)
		require.NotNil(t, profile)
		return profile, ScheduleOrigin(profile.Name, constants.CommandBackup)
	}

	t.Run("create-united-with-nil", func(t *testing.T) {
		schedule := NewSchedule(nil, nil)
		assert.NotEqual(t, scheduleBaseConfigDefaults, schedule.ScheduleBaseConfig)
	})

	t.Run("default-can-init-with-nil", func(t *testing.T) {
		origin := ScheduleOrigin("p", "c")
		schedule := NewDefaultSchedule(nil, origin)
		assert.Equal(t, origin, schedule.ScheduleOrigin())
		assert.Equal(t, scheduleBaseConfigDefaults, schedule.ScheduleBaseConfig)
	})

	t.Run("default-without-schedule", func(t *testing.T) {
		// Ensure DefaultSchedule can be used as remove-only config
		p, origin := profile(t, ``)
		schedule := NewDefaultSchedule(p.config, origin)
		assert.False(t, schedule.HasSchedules())
	})

	t.Run("global defaults", func(t *testing.T) {
		p, origin := profile(t, `
			[global]
			systemd-drop-in-files = "drop-in-file.conf"

			[global.schedule-defaults]
			log = "global-custom.log"
			lock-wait = "30s"

			[default.backup]
			schedule = "daily"
		`)
		t.Run("schedule-defaults apply", func(t *testing.T) {
			for i := 0; i < 2; i++ {
				var schedule *Schedule
				if i == 0 {
					schedule = NewDefaultSchedule(p.config, origin)
				} else {
					schedule = p.Schedules()["backup"]
				}
				assert.Equal(t, "global-custom.log", schedule.Log)
				assert.Equal(t, 30*time.Second, schedule.GetLockWait())
				assert.Equal(t, []string{"drop-in-file.conf"}, schedule.SystemdDropInFiles)
			}
		})
		t.Run("schedule-defaults do not apply", func(t *testing.T) {
			schedule := NewSchedule(p.config, NewDefaultScheduleConfig(nil, origin))
			assert.Empty(t, schedule.Log)
			assert.Equal(t, 0*time.Second, schedule.GetLockWait())
			// other global defaults are applied
			assert.Equal(t, []string{"drop-in-file.conf"}, schedule.SystemdDropInFiles)
		})
	})

	t.Run("profile schedule overrides", func(t *testing.T) {
		p, _ := profile(t, `
			[default]
			systemd-drop-in-files = "my-systemd-drop-in.conf"

			[default.backup]
			schedule-log = "overridden.log"
			schedule-lock-wait = "55s"

			[default.backup.schedule]
			at = "monthly"
			log = "schedule.log"
			lock-mode = "ignore"
			lock-wait = "30s"
		`)

		schedule := p.Schedules()["backup"]
		assert.Equal(t, []string{"monthly"}, schedule.Schedules)
		assert.Equal(t, []string{"my-systemd-drop-in.conf"}, schedule.SystemdDropInFiles)
		assert.Equal(t, "overridden.log", schedule.Log)
		assert.Equal(t, 55*time.Second, schedule.GetLockWait())
		assert.Equal(t, "ignore", schedule.LockMode)
	})

	t.Run("profile inline schedule", func(t *testing.T) {
		p, _ := profile(t, `
			[default.backup]
			schedule = ["10:00", "weekly"]

			[default.check]
			schedule = "daily"
		`)

		schedule := p.Schedules()["backup"]
		assert.Equal(t, []string{"10:00", "weekly"}, schedule.Schedules)
		schedule = p.Schedules()["check"]
		assert.Equal(t, []string{"daily"}, schedule.Schedules)
	})

	t.Run("profile environment", func(t *testing.T) {
		p, _ := profile(t, `
			[default.env]
			MY_KEY = "value"
			MY_PASSWORD = "plain"
			OTHER_KEY = "value"
			RESTICPROFILE_SCHEDULE_ID = "cannot-override"

			[default.backup.schedule]
			at = "daily"
			capture-environment = "MY_*"
		`)

		schedule := p.Schedules()["backup"]
		assert.Equal(t, []string{
			"MY_KEY=value",
			"MY_PASSWORD=plain",
			"RESTICPROFILE_SCHEDULE_ID=:backup@default",
		}, schedule.Environment)
	})
}

func TestQueryNilScheduleConfig(t *testing.T) {
	var config *ScheduleConfig
	assert.False(t, config.HasSchedules())
}

func TestScheduleBuiltinDefaults(t *testing.T) {
	s := NewDefaultSchedule(nil, ScheduleOrigin("", ""))
	require.Equal(t, scheduleBaseConfigDefaults, s.ScheduleBaseConfig)

	assert.Equal(t, "auto", s.Permission)
	assert.Equal(t, "background", s.Priority)
	assert.Equal(t, "default", s.LockMode)
	assert.Equal(t, []string{"RESTIC_*"}, s.EnvCapture)
	assert.Equal(t, ScheduleLockModeDefault, s.GetLockMode())
}

func TestLockModes(t *testing.T) {
	tests := map[ScheduleLockMode]ScheduleBaseConfig{
		ScheduleLockModeDefault: {LockMode: ""},
		ScheduleLockModeFail:    {LockMode: constants.ScheduleLockModeOptionFail},
		ScheduleLockModeIgnore:  {LockMode: constants.ScheduleLockModeOptionIgnore},
	}
	for mode, config := range tests {
		s := Schedule{}
		s.ScheduleBaseConfig = config
		assert.Equal(t, mode, s.GetLockMode())
	}
}

func TestLockWait(t *testing.T) {
	tests := map[time.Duration]ScheduleBaseConfig{
		0:               {LockWait: maybe.SetDuration(2 * time.Second)}, // min lock wait is >2 seconds
		3 * time.Second: {LockWait: maybe.SetDuration(3 * time.Second)},
		120 * time.Hour: {LockWait: maybe.SetDuration(120 * time.Hour)},
	}
	for mode, config := range tests {
		s := Schedule{}
		s.ScheduleBaseConfig = config
		assert.Equal(t, mode, s.GetLockWait())
	}
}

func TestScheduleFlags(t *testing.T) {
	schedule := &Schedule{}

	flag, found := schedule.GetFlag("unit")
	assert.Empty(t, flag)
	assert.False(t, found)

	schedule.SetFlag("unit", "test")
	flag, found = schedule.GetFlag("unit")
	assert.Equal(t, "test", flag)
	assert.True(t, found)
}
