package config

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchedule(t *testing.T) {
	profile := func(t *testing.T, config string) (*Profile, ScheduleConfigOrigin) {
		t.Helper()
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

	t.Run("profile origin", func(t *testing.T) {
		// Ensure DefaultSchedule can be used as remove-only config
		p, origin := profile(t, `
			[default.backup]
			schedule = "daily"
		`)
		assert.Equal(t, origin, p.Schedules()["backup"].ScheduleOrigin())
	})

	t.Run("global defaults", func(t *testing.T) {
		p, origin := profile(t, `
			[global.schedule-defaults]
			systemd-drop-in-files = "drop-in-file.conf"
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
			assert.Empty(t, schedule.SystemdDropInFiles)
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

	t.Run("profile drop-in overrides", func(t *testing.T) {
		p, _ := profile(t, `
			[global.schedule-defaults]
			systemd-drop-in-files = "drop-in-file.conf"
			
			[default]
			systemd-drop-in-files = "default-drop-in.conf"
			
			[default.backup.schedule]
			at = "monthly"
			
			[default.check.schedule]
			at = "monthly"
			systemd-drop-in-files = "check-drop-in.conf"
		`)

		assert.Equal(t, []string{"default-drop-in.conf"}, p.Schedules()["backup"].SystemdDropInFiles)
		assert.Equal(t, []string{"check-drop-in.conf"}, p.Schedules()["check"].SystemdDropInFiles)
	})

	t.Run("profile schedule parse error", func(t *testing.T) {
		defaultLogger := clog.GetDefaultLogger()
		defer clog.SetDefaultLogger(defaultLogger)
		mem := clog.NewMemoryHandler()
		clog.SetDefaultLogger(clog.NewLogger(mem))

		p, _ := profile(t, `
			[default.backup.schedule]
			at = true
			non-existing = "xyz"

			[default.check.schedule]
			at = "daily"
			lock-wait = ["invalid"]
		`)

		schedule := p.Schedules()["backup"]
		assert.Equal(t, []string{"1"}, schedule.Schedules)

		assert.Nil(t, p.Schedules()["check"])
		msg := `failed decoding schedule {"at":"daily","lock-wait":["invalid"]}: 1 error(s) decoding:

* 'lock-wait' expected a map, got 'slice'`
		assert.Contains(t, mem.Logs(), msg)
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

	t.Run("profile undefined schedule", func(t *testing.T) {
		p, _ := profile(t, `
			[default.backup]
			schedule = ""

			[default.check.schedule]
			at = ""
		`)

		assert.Nil(t, p.Schedules()["backup"])
		assert.Nil(t, p.Schedules()["check"])
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

func TestNewScheduleFromGroup(t *testing.T) {
	group := func(t *testing.T, config string) (*Group, ScheduleConfigOrigin) {
		t.Helper()
		config = "version = \"2\"\n\n" + config
		if !strings.Contains(config, "profiles =") {
			config += `
			[groups.default]
			profiles = "default"
			`
		}
		c, err := Load(bytes.NewBufferString(config), "toml")
		require.NoError(t, err)
		group, err := c.GetProfileGroup("default")
		require.NoError(t, err)
		require.NotNil(t, group)
		return group, ScheduleOrigin(group.Name, constants.CommandBackup, ScheduleOriginGroup)
	}

	t.Run("group without schedule", func(t *testing.T) {
		g, _ := group(t, ``)
		assert.Empty(t, g.Schedules())
	})

	t.Run("group with undefined schedule", func(t *testing.T) {
		g, _ := group(t, `
			[groups.default.schedules.backup]
			at = "" # empty schedule
		`)
		assert.Empty(t, g.Schedules())
	})

	t.Run("group with schedule", func(t *testing.T) {
		g, _ := group(t, `
			[groups.default.schedules.backup]
			at = "daily"
			log = "group-backup.log"
			[groups.default.schedules.check]
			at = "monthly"
			log = "group-check.log"
		`)
		require.Len(t, g.Schedules(), 2)
		backup, check := g.Schedules()["backup"], g.Schedules()["check"]
		require.NotNil(t, backup)
		require.Equal(t, []string{"daily"}, backup.Schedules)
		require.Equal(t, "group-backup.log", backup.Log)
		require.Equal(t, []string{"monthly"}, check.Schedules)
		require.Equal(t, "group-check.log", check.Log)
	})

	t.Run("group origin", func(t *testing.T) {
		g, origin := group(t, `
			[groups.default.schedules.backup]
			at = "daily"
		`)
		assert.Equal(t, origin, g.Schedules()["backup"].ScheduleOrigin())
	})

	t.Run("global defaults", func(t *testing.T) {
		g, origin := group(t, `
			[global.schedule-defaults]
			systemd-drop-in-files = "drop-in-file.conf"
			log = "global-custom.log"
			lock-wait = "30s"

			[groups.default.schedules.backup]
			at = "daily"
		`)
		t.Run("schedule-defaults apply", func(t *testing.T) {
			for i := 0; i < 2; i++ {
				var schedule *Schedule
				if i == 0 {
					schedule = NewDefaultSchedule(g.config, origin)
				} else {
					schedule = g.Schedules()["backup"]
					assert.Equal(t, []string{"daily"}, schedule.Schedules)
				}
				assert.Equal(t, "global-custom.log", schedule.Log)
				assert.Equal(t, 30*time.Second, schedule.GetLockWait())
				assert.Equal(t, []string{"drop-in-file.conf"}, schedule.SystemdDropInFiles)
			}
		})
	})
}

func TestQueryNilScheduleConfig(t *testing.T) {
	var config *ScheduleConfig
	assert.False(t, config.HasSchedules())
}

func TestNormalizeLogPath(t *testing.T) {
	sep := regexp.MustCompile(`[/\\]`)
	baseDir, _ := filepath.Abs(t.TempDir())
	s := NewSchedule(nil, NewDefaultScheduleConfig(nil, ScheduleOrigin("", "")))
	s.ConfigFile = filepath.Join(baseDir, "profiles.yaml")

	expected := filepath.Join(baseDir, "schedule.log")
	s.Log = "schedule.log"
	s.init(nil)
	assert.Equal(t, sep.Split(expected, -1), sep.Split(s.Log, -1))

	expected = "tcp://localhost"
	s.Log = "tcp://localhost"
	s.init(nil)
	assert.Equal(t, expected, s.Log)
}

func TestCompareSchedules(t *testing.T) {
	cfgA := NewDefaultScheduleConfig(nil, ScheduleOrigin("a-name", "a-command"))
	cfgB := NewDefaultScheduleConfig(nil, ScheduleOrigin("b-name", "b-command"))
	cfgC := NewDefaultScheduleConfig(nil, ScheduleOrigin("a-name", "b-command"))
	a, b, c := NewSchedule(nil, cfgA), NewSchedule(nil, cfgB), NewSchedule(nil, cfgC)

	assert.Equal(t, 0, CompareSchedules(nil, nil))
	assert.Equal(t, 0, CompareSchedules(a, a))
	assert.Equal(t, 0, CompareSchedules(b, b))
	assert.Equal(t, 1, CompareSchedules(a, nil))
	assert.Equal(t, -1, CompareSchedules(nil, a))
	assert.Equal(t, -1, CompareSchedules(a, b))
	assert.Equal(t, 1, CompareSchedules(b, a))
	assert.Equal(t, 1, CompareSchedules(c, a))
	assert.Equal(t, -1, CompareSchedules(a, c))
}

func TestScheduleForProfileEnforcesOrigin(t *testing.T) {
	profile := NewProfile(nil, "profile")
	config := NewDefaultScheduleConfig(nil, ScheduleOrigin(profile.Name, "backup", ScheduleOriginGroup))

	assert.PanicsWithError(t, "invalid use of newScheduleForProfile(profile, g:backup@profile)", func() {
		newScheduleForProfile(profile, config)
	})

	config.origin.Type = ScheduleOriginProfile
	assert.NotNil(t, newScheduleForProfile(profile, config))
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
