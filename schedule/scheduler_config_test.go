package schedule

import (
	"fmt"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestOsDefaultConfig(t *testing.T) {
	g := &config.Global{SystemdTimerTemplate: "timer.tpl", SystemdUnitTemplate: "unit.tpl"}
	cfg := NewSchedulerConfig(g)

	t.Run("is-os-default", func(t *testing.T) {
		assert.Equal(t, constants.SchedulerOSDefault, cfg.Type())
	})

	t.Run("can-convert-to-systemd", func(t *testing.T) {
		systemd := cfg.Convert(constants.SchedulerSystemd)
		assert.Equal(t, SchedulerSystemd{TimerTemplate: "timer.tpl", UnitTemplate: "unit.tpl"}, systemd)
	})
}

func TestSystemdConfig(t *testing.T) {
	testCases := []struct {
		global   *config.Global
		expected SchedulerSystemd
	}{
		{
			global: &config.Global{
				Scheduler:            constants.SchedulerSystemd,
				SystemdTimerTemplate: "timer.tpl",
				SystemdUnitTemplate:  "unit.tpl",
			},
			expected: SchedulerSystemd{TimerTemplate: "timer.tpl", UnitTemplate: "unit.tpl"},
		},
		{
			global: &config.Global{
				Scheduler:            constants.SchedulerSystemd,
				SystemdTimerTemplate: "timer.tpl",
				SystemdUnitTemplate:  "unit.tpl",
				IONiceClass:          3,
				IONiceLevel:          5,
			},
			expected: SchedulerSystemd{TimerTemplate: "timer.tpl", UnitTemplate: "unit.tpl"},
		},
		{
			global: &config.Global{
				Scheduler:            constants.SchedulerSystemd,
				SystemdTimerTemplate: "timer.tpl",
				SystemdUnitTemplate:  "unit.tpl",
				IONice:               true,
				IONiceClass:          3,
				IONiceLevel:          5,
			},
			expected: SchedulerSystemd{TimerTemplate: "timer.tpl", UnitTemplate: "unit.tpl", IONiceClass: 3, IONiceLevel: 5},
		},
		{
			global: &config.Global{
				Scheduler:            constants.SchedulerSystemd,
				SystemdTimerTemplate: "timer.tpl",
				SystemdUnitTemplate:  "unit.tpl",
				Nice:                 12,
			},
			expected: SchedulerSystemd{TimerTemplate: "timer.tpl", UnitTemplate: "unit.tpl", Nice: 12},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, NewSchedulerConfig(tc.global))
	}
}

func TestCrondConfig(t *testing.T) {
	global := func(res string) *config.Global {
		return &config.Global{Scheduler: fmt.Sprintf("%s : %s", constants.SchedulerCrond, res)}
	}
	t.Run("default-binary", func(t *testing.T) {
		assert.Equal(t, SchedulerCrond{}, NewSchedulerConfig(global("")))
	})
	t.Run("custom-binary", func(t *testing.T) {
		assert.Equal(t, SchedulerCrond{CrontabBinary: "/my/binary"}, NewSchedulerConfig(global("/my/binary")))
	})
}

func TestCrontabConfig(t *testing.T) {
	global := func(res string) *config.Global {
		return &config.Global{Scheduler: fmt.Sprintf("%s : %s", constants.SchedulerCrontab, res)}
	}
	t.Run("no-file-panic", func(t *testing.T) {
		msg := `invalid schedule "crontab", no crontab file was specified, expecting "crontab: filename"`
		assert.PanicsWithError(t, msg, func() { NewSchedulerConfig(global("")) })
	})
	t.Run("file-detect-user", func(t *testing.T) {
		assert.Equal(t, SchedulerCrond{CrontabFile: "/my/file"}, NewSchedulerConfig(global("/my/file")))
	})
	t.Run("file-with-user", func(t *testing.T) {
		assert.Equal(t, SchedulerCrond{CrontabFile: "/my/file", Username: "user"}, NewSchedulerConfig(global("user : /my/file")))
	})
	t.Run("file-no-user", func(t *testing.T) {
		assert.Equal(t, SchedulerCrond{CrontabFile: "/my/file", Username: "-"}, NewSchedulerConfig(global(" : /my/file")))
	})
}
