package config

import (
	"time"

	"github.com/creativeprojects/resticprofile/constants"
)

// Global holds the configuration from the global section
type Global struct {
	IONice               bool          `mapstructure:"ionice" default:"false" description:"Enables setting the unix IO priority class and level for resticprofile and child processes (only on unix OS)."`
	IONiceClass          int           `mapstructure:"ionice-class" default:"2" range:"[1:3]" description:"Sets the unix \"ionice-class\" to apply when \"ionice\" is enabled"`
	IONiceLevel          int           `mapstructure:"ionice-level" default:"0" range:"[0:7]" description:"Sets the unix \"ionice-level\" to apply when \"ionice\" is enabled"`
	Nice                 int           `mapstructure:"nice" default:"0" range:"[-20:19]" description:"Sets the unix \"nice\" value for resticprofile and child processes (on any OS)"`
	Priority             string        `mapstructure:"priority" default:"normal" enum:"idle;background;low;normal;high;highest" description:"Sets process priority class for resticprofile and child processes (on any OS)"`
	DefaultCommand       string        `mapstructure:"default-command" default:"snapshots" description:"The restic or resticprofile command to use when no command was specified"`
	Initialize           bool          `mapstructure:"initialize" default:"false" description:"Initialize a repository if missing"`
	ResticBinary         string        `mapstructure:"restic-binary" description:"Full path of the restic executable (detected if not set)"`
	ResticVersion        string        // not configurable at the moment. To be set after ResticBinary is known.
	FilterResticFlags    bool          `mapstructure:"restic-arguments-filter" default:"true" description:"Remove unknown flags instead of passing all configured flags to restic"`
	ResticLockRetryAfter time.Duration `mapstructure:"restic-lock-retry-after" default:"1m" description:"Time to wait before trying to get a lock on a restic repositoey - see https://creativeprojects.github.io/resticprofile/usage/locks/"`
	ResticStaleLockAge   time.Duration `mapstructure:"restic-stale-lock-age" default:"1h" description:"The age an unused lock on a restic repository must have at least before resiticprofile attempts to unlock - see https://creativeprojects.github.io/resticprofile/usage/locks/"`
	ShellBinary          []string      `mapstructure:"shell" default:"auto" examples:"sh;bash;pwsh;powershell;cmd" description:"The shell that is used to run commands (default is OS specific)"`
	MinMemory            uint64        `mapstructure:"min-memory" default:"100" description:"Minimum available memory (in MB) required to run any commands - see https://creativeprojects.github.io/resticprofile/usage/memory/"`
	Scheduler            string        `mapstructure:"scheduler" description:"Leave blank for the default scheduler or use \"crond\" to select cron on supported operating systems"`
	LegacyArguments      bool          `mapstructure:"legacy-arguments" default:"false" deprecated:"0.20.0" description:"Legacy, broken arguments mode of resticprofile before version 0.15"`
	SystemdUnitTemplate  string        `mapstructure:"systemd-unit-template" default:"" description:"File containing the go template to generate a systemd unit - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	SystemdTimerTemplate string        `mapstructure:"systemd-timer-template" default:"" description:"File containing the go template to generate a systemd timer - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	SystemdDropInFiles   []string      `mapstructure:"systemd-drop-in-files" default:"" description:"Files containing systemd drop-in (override) files - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	SenderTimeout        time.Duration `mapstructure:"send-timeout" default:"30s" examples:"15s;30s;2m30s" description:"Timeout when sending messages to a webhook - see https://creativeprojects.github.io/resticprofile/configuration/http_hooks/"`
	CACertificates       []string      `mapstructure:"ca-certificates" description:"Path to PEM encoded certificates to trust in addition to system certificates when resticprofile sends to a webhook - see https://creativeprojects.github.io/resticprofile/configuration/http_hooks/"`
	PreventSleep         bool          `mapstructure:"prevent-sleep" default:"false" description:"Prevent the system from sleeping while running commands - see https://creativeprojects.github.io/resticprofile/configuration/sleep/"`
	GroupContinueOnError bool          `mapstructure:"group-continue-on-error" default:"false" description:"Enable groups to continue with the next profile(s) instead of stopping at the first failure"`
}

// NewGlobal instantiates a new Global with default values
func NewGlobal() *Global {
	return &Global{
		IONice:               constants.DefaultIONiceFlag,
		IONiceClass:          constants.DefaultIONiceClass,
		Nice:                 constants.DefaultStandardNiceFlag,
		DefaultCommand:       constants.DefaultCommand,
		FilterResticFlags:    constants.DefaultFilterResticFlags,
		ResticLockRetryAfter: constants.DefaultResticLockRetryAfter,
		ResticStaleLockAge:   constants.DefaultResticStaleLockAge,
		MinMemory:            constants.DefaultMinMemory,
		SenderTimeout:        constants.DefaultSenderTimeout,
	}
}

func (p *Global) SetRootPath(rootPath string) {
	p.ShellBinary = fixPaths(p.ShellBinary, expandEnv)
	p.ResticBinary = fixPath(p.ResticBinary, expandEnv)

	p.SystemdUnitTemplate = fixPath(p.SystemdUnitTemplate, expandEnv, absolutePrefix(rootPath))
	p.SystemdTimerTemplate = fixPath(p.SystemdTimerTemplate, expandEnv, absolutePrefix(rootPath))
	p.SystemdDropInFiles = fixPaths(p.SystemdDropInFiles, expandEnv, absolutePrefix(rootPath))

	for index, file := range p.CACertificates {
		p.CACertificates[index] = fixPath(file, expandEnv, absolutePrefix(rootPath))
	}
}
