package config

import (
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
)

// Global holds the configuration from the global section
type Global struct {
	IONice               bool                `mapstructure:"ionice" default:"false" description:"Enables setting the linux IO priority class and level for resticprofile and child processes (only on linux OS)."`
	IONiceClass          int                 `mapstructure:"ionice-class" default:"2" range:"[1:3]" description:"Sets the linux \"ionice-class\" (I/O scheduling class) to apply when \"ionice\" is enabled (1=realtime, 2=best-effort, 3=idle)"`
	IONiceLevel          int                 `mapstructure:"ionice-level" default:"0" range:"[0:7]" description:"Sets the linux \"ionice-level\" (I/O priority within the scheduling class) to apply when \"ionice\" is enabled (0=highest priority, 7=lowest priority)"`
	Nice                 int                 `mapstructure:"nice" default:"0" range:"[-20:19]" description:"Sets the unix \"nice\" value for resticprofile and child processes (on any OS)"`
	Priority             string              `mapstructure:"priority" default:"normal" enum:"idle;background;low;normal;high;highest" description:"Sets process priority class for resticprofile and child processes (on any OS)"`
	DefaultCommand       string              `mapstructure:"default-command" default:"snapshots" description:"The restic or resticprofile command to use when no command was specified"`
	Initialize           bool                `mapstructure:"initialize" default:"false" description:"Initialize a repository if missing"`
	NoAutoRepositoryFile maybe.Bool          `mapstructure:"prevent-auto-repository-file" default:"false" description:"Prevents using a repository file for repository definitions containing a password"`
	ResticBinary         string              `mapstructure:"restic-binary" description:"Full path of the restic executable (detected if not set)"`
	ResticVersion        string              `mapstructure:"restic-version" pattern:"^(|[0-9]+\\.[0-9]+(\\.[0-9]+)?)$" description:"Sets the restic version (detected if not set)"`
	FilterResticFlags    bool                `mapstructure:"restic-arguments-filter" default:"true" description:"Remove unknown flags instead of passing all configured flags to restic"`
	ResticLockRetryAfter time.Duration       `mapstructure:"restic-lock-retry-after" default:"1m" description:"Time to wait before trying to get a lock on a restic repository - see https://creativeprojects.github.io/resticprofile/usage/locks/"`
	ResticStaleLockAge   time.Duration       `mapstructure:"restic-stale-lock-age" default:"1h" description:"The age an unused lock on a restic repository must have at least before resticprofile attempts to unlock - see https://creativeprojects.github.io/resticprofile/usage/locks/"`
	ShellBinary          []string            `mapstructure:"shell" default:"auto" examples:"sh;bash;pwsh;powershell;cmd" description:"The shell that is used to run commands (default is OS specific)"`
	MinMemory            uint64              `mapstructure:"min-memory" default:"100" description:"Minimum available memory (in MB) required to run any commands - see https://creativeprojects.github.io/resticprofile/usage/memory/"`
	Scheduler            string              `mapstructure:"scheduler" default:"auto" examples:"auto;launchd;systemd;taskscheduler;crond;crond:/usr/bin/crontab;crontab:*:/etc/cron.d/resticprofile" description:"Selects the scheduler. Blank or \"auto\" uses the default scheduler of your operating system: \"launchd\", \"systemd\", \"taskscheduler\" or \"crond\" (as fallback). Alternatively you can set \"crond\" for cron compatible schedulers supporting the crontab executable API or \"crontab:[user:]file\" to write into a crontab file directly. The need for a user is detected if missing and can be set to a name, \"-\" (no user) or \"*\" (current user)."`
	ScheduleDefaults     *ScheduleBaseConfig `mapstructure:"schedule-defaults" default:"" description:"Sets defaults for all schedules"`
	Log                  string              `mapstructure:"log" default:"" examples:"/resticprofile.log;syslog-tcp://syslog-server:514;syslog:server;syslog:" description:"Sets the default log destination to be used if not specified in \"--log\" or \"schedule-log\" - see https://creativeprojects.github.io/resticprofile/configuration/logs/"`
	CommandOutput        string              `mapstructure:"command-output" default:"auto" enum:"auto;log;console;all" description:"Sets the destination for command output (stderr/stdout). \"log\" sends output to the log file (if specified), \"console\" sends it to the console instead. \"auto\" sends it to \"both\" if console is a terminal otherwise to \"log\" only - see https://creativeprojects.github.io/resticprofile/configuration/logs/"`
	LegacyArguments      bool                `mapstructure:"legacy-arguments" default:"false" deprecated:"0.20.0" description:"Legacy, broken arguments mode of resticprofile before version 0.15"`
	SystemdUnitTemplate  string              `mapstructure:"systemd-unit-template" default:"" description:"File containing the go template to generate a systemd unit - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	SystemdTimerTemplate string              `mapstructure:"systemd-timer-template" default:"" description:"File containing the go template to generate a systemd timer - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	SenderTimeout        time.Duration       `mapstructure:"send-timeout" default:"30s" examples:"15s;30s;2m30s" description:"Timeout when sending messages to a webhook - see https://creativeprojects.github.io/resticprofile/configuration/http_hooks/"`
	CACertificates       []string            `mapstructure:"ca-certificates" description:"Path to PEM encoded certificates to trust in addition to system certificates when resticprofile sends to a webhook - see https://creativeprojects.github.io/resticprofile/configuration/http_hooks/"`
	PreventSleep         bool                `mapstructure:"prevent-sleep" default:"false" description:"Prevent the system from sleeping while running commands - see https://creativeprojects.github.io/resticprofile/configuration/sleep/"`
	GroupContinueOnError bool                `mapstructure:"group-continue-on-error" default:"false" description:"Enable groups to continue with the next profile(s) instead of stopping at the first failure"`
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
		CommandOutput:        constants.DefaultCommandOutput,
		SenderTimeout:        constants.DefaultSenderTimeout,
	}
}

func (p *Global) SetRootPath(rootPath string) {
	p.ShellBinary = fixPaths(p.ShellBinary, expandEnv)
	p.ResticBinary = fixPath(p.ResticBinary, expandEnv)
	p.Log = fixPath(p.Log, expandEnv, expandUserHome)

	p.SystemdUnitTemplate = fixPath(p.SystemdUnitTemplate, expandEnv, absolutePrefix(rootPath))
	p.SystemdTimerTemplate = fixPath(p.SystemdTimerTemplate, expandEnv, absolutePrefix(rootPath))

	for index, file := range p.CACertificates {
		p.CACertificates[index] = fixPath(file, expandEnv, absolutePrefix(rootPath))
	}
}
