package config

import (
	"time"

	"github.com/creativeprojects/resticprofile/constants"
)

// Global holds the configuration from the global section
type Global struct {
	IONice               bool          `mapstructure:"ionice"`
	IONiceClass          int           `mapstructure:"ionice-class"`
	IONiceLevel          int           `mapstructure:"ionice-level"`
	Nice                 int           `mapstructure:"nice"`
	Priority             string        `mapstructure:"priority"`
	DefaultCommand       string        `mapstructure:"default-command"`
	Initialize           bool          `mapstructure:"initialize"`
	ResticBinary         string        `mapstructure:"restic-binary"`
	ResticLockRetryAfter time.Duration `mapstructure:"restic-lock-retry-after"`
	ResticStaleLockAge   time.Duration `mapstructure:"restic-stale-lock-age"`
	ShellBinary          []string      `mapstructure:"shell"`
	MinMemory            uint64        `mapstructure:"min-memory"`
	Scheduler            string        `mapstructure:"scheduler"`
	LegacyArguments      bool          `mapstructure:"legacy-arguments"` // broken arguments mode (before v0.15.0)
	SystemdUnitTemplate  string        `mapstructure:"systemd-unit-template"`
	SystemdTimerTemplate string        `mapstructure:"systemd-timer-template"`
	SenderTimeout        time.Duration `mapstructure:"send-timeout"`
	CACertificates       []string      `mapstructure:"ca-certificates"`
	PreventSleep         bool          `mapstructure:"prevent-sleep"`
	GroupNextOnError     bool          `mapstructure:"group-next-on-error"`
}

// NewGlobal instantiates a new Global with default values
func NewGlobal() *Global {
	return &Global{
		IONice:               constants.DefaultIONiceFlag,
		Nice:                 constants.DefaultStandardNiceFlag,
		DefaultCommand:       constants.DefaultCommand,
		ResticLockRetryAfter: constants.DefaultResticLockRetryAfter,
		ResticStaleLockAge:   constants.DefaultResticStaleLockAge,
		MinMemory:            constants.DefaultMinMemory,
		SenderTimeout:        constants.DefaultSenderTimeout,
	}
}

func (p *Global) SetRootPath(rootPath string) {
	p.SystemdUnitTemplate = fixPath(p.SystemdUnitTemplate, expandEnv, absolutePrefix(rootPath))
	p.SystemdTimerTemplate = fixPath(p.SystemdTimerTemplate, expandEnv, absolutePrefix(rootPath))

	if len(p.CACertificates) == 0 {
		return
	}
	for index, file := range p.CACertificates {
		p.CACertificates[index] = fixPath(file, expandEnv, absolutePrefix(rootPath))
	}
}
