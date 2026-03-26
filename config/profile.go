package config

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/mitchellh/mapstructure"
)

// resticVersion14 is the semver of restic 0.14 (the version where several flag names were changed)
var resticVersion14 = semver.MustParse("0.14")

// Empty allows to test if a section is specified or not
type Empty interface {
	IsEmpty() bool
}

// Monitoring provides access to http hooks inside a section
type Monitoring interface {
	GetSendMonitoring() *SendMonitoringSections
}

// RunShellCommands provides access to shell command hooks inside a section
type RunShellCommands interface {
	GetRunShellCommands() *RunShellCommandsSection
}

// OtherFlags provides access to dynamic commandline flags
type OtherFlags interface {
	GetOtherFlags() map[string]any
}

// scheduling provides access to schedule information inside a section
type scheduling interface {
	getScheduleConfig(p *Profile, command string) *ScheduleConfig
}

// commandFlags allows sections to return flags directly
type commandFlags interface {
	getCommandFlags(profile *Profile) *shell.Args
}

// relativePath allows sections to take part in Profile.SetRootPath
type relativePath interface {
	setRootPath(profile *Profile, rootPath string)
}

// resolver allows sections to take part in Profile.ResolveConfiguration
type resolver interface {
	resolve(profile *Profile)
}

// Profile contains the whole profile configuration
type Profile struct {
	RunShellCommandsSection `mapstructure:",squash"`
	OtherFlagsSection       `mapstructure:",squash"`

	config               *Config
	resticVersion        *semver.Version
	Name                 string
	Description          string                       `mapstructure:"description" description:"Describes the profile"`
	BaseDir              string                       `mapstructure:"base-dir" description:"Sets the working directory for this profile. The profile will fail when the working directory cannot be changed. Leave empty to use the current directory instead"`
	Quiet                bool                         `mapstructure:"quiet" argument:"quiet"`
	Verbose              int                          `mapstructure:"verbose" argument:"verbose"`
	KeyHint              string                       `mapstructure:"key-hint" argument:"key-hint"`
	Repository           ConfidentialValue            `mapstructure:"repository" argument:"repo"`
	RepositoryFile       string                       `mapstructure:"repository-file" argument:"repository-file"`
	PasswordFile         string                       `mapstructure:"password-file" argument:"password-file"`
	PasswordCommand      string                       `mapstructure:"password-command" argument:"password-command"`
	CacheDir             string                       `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert               string                       `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert        string                       `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize           bool                         `mapstructure:"initialize" default:"" description:"Initialize the restic repository if missing"`
	Inherit              string                       `mapstructure:"inherit" show:"noshow" description:"Name of the profile to inherit all of the settings from"`
	Lock                 string                       `mapstructure:"lock" description:"Path to the lock file to use with resticprofile locks"`
	ForceLock            bool                         `mapstructure:"force-inactive-lock" description:"Allows to lock when the existing lock is considered stale"`
	StreamError          []StreamErrorSection         `mapstructure:"stream-error" description:"Run shell command(s) when a pattern matches the stderr of restic"`
	StatusFile           string                       `mapstructure:"status-file" description:"Path to the status file to update with a summary of last restic command result"`
	PrometheusSaveToFile string                       `mapstructure:"prometheus-save-to-file" description:"Path to the prometheus metrics file to update with a summary of the last restic command result"`
	PrometheusPush       string                       `mapstructure:"prometheus-push" format:"uri" description:"URL of the prometheus push gateway to send the summary of the last restic command result to"`
	PrometheusPushJob    string                       `mapstructure:"prometheus-push-job" description:"Prometheus push gateway job name. $command placeholder is replaced with restic command"`
	PrometheusPushFormat string                       `mapstructure:"prometheus-push-format" default:"text" enum:"text;protobuf" description:"Prometheus push gateway request format"`
	PrometheusLabels     map[string]string            `mapstructure:"prometheus-labels" description:"Additional prometheus labels to set"`
	SystemdDropInFiles   []string                     `mapstructure:"systemd-drop-in-files" default:"" description:"Files containing systemd drop-in (override) files - see https://creativeprojects.github.io/resticprofile/schedules/systemd/"`
	Environment          map[string]ConfidentialValue `mapstructure:"env" description:"Additional environment variables to set in any child process. Inline env variables take precedence over dotenv files declared with \"env-file\"."`
	EnvironmentFiles     []string                     `mapstructure:"env-file" description:"Additional dotenv files to load and set as environment in any child process"`
	Init                 *InitSection                 `mapstructure:"init"`
	Backup               *BackupSection               `mapstructure:"backup"`
	Retention            *RetentionSection            `mapstructure:"retention" command:"forget"`
	Check                *GenericSectionWithSchedule  `mapstructure:"check"`
	Prune                *GenericSectionWithSchedule  `mapstructure:"prune"`
	Forget               *GenericSectionWithSchedule  `mapstructure:"forget"`
	Copy                 *CopySection                 `mapstructure:"copy"`
	OtherSections        map[string]*GenericSection   `show:",remain"`
}

// GenericSection is used for all restic commands that are not covered in specific section types
type GenericSection struct {
	OtherFlagsSection       `mapstructure:",squash"`
	RunShellCommandsSection `mapstructure:",squash"`
	SendMonitoringSections  `mapstructure:",squash"`
}

func (g *GenericSection) setRootPath(p *Profile, rootPath string) {
	g.SendMonitoringSections.setRootPath(p, rootPath)
}

func (g *GenericSection) IsEmpty() bool { return g == nil }

// InitSection contains the specific configuration to the 'init' command
type InitSection struct {
	GenericSection `mapstructure:",squash"`

	CopyChunkerParams   bool              `mapstructure:"copy-chunker-params" argument:"copy-chunker-params"`
	FromKeyHint         string            `mapstructure:"from-key-hint" argument:"from-key-hint"`
	FromRepository      ConfidentialValue `mapstructure:"from-repository" argument:"from-repo"`
	FromRepositoryFile  string            `mapstructure:"from-repository-file" argument:"from-repository-file"`
	FromPasswordFile    string            `mapstructure:"from-password-file" argument:"from-password-file"`
	FromPasswordCommand string            `mapstructure:"from-password-command" argument:"from-password-command"`
}

func (i *InitSection) IsEmpty() bool { return i == nil }

func (i *InitSection) resolve(_ *Profile) {
	i.FromRepository.setValue(fixPath(i.FromRepository.Value(), expandEnv, expandUserHome))
}

func (i *InitSection) setRootPath(p *Profile, rootPath string) {
	i.GenericSection.setRootPath(p, rootPath)

	i.FromRepositoryFile = fixPath(i.FromRepositoryFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	i.FromPasswordFile = fixPath(i.FromPasswordFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
}

func (i *InitSection) getCommandFlags(profile *Profile) (flags *shell.Args) {
	legacyArgs := map[string]string{
		"from-repo":             "repo2",
		"from-repository-file":  "repository-file2",
		"from-password-file":    "password-file2",
		"from-password-command": "password-command2",
		"from-key-hint":         "key-hint2",
	}

	// Handle confidential repo in flags
	restore := profile.replaceWithRepositoryFile(&i.FromRepository, &i.FromRepositoryFile, "-from")
	defer restore()

	flags = profile.GetCommonFlags()
	addArgsFromStruct(flags, i)
	addArgsFromOtherFlags(flags, profile, i)

	if v := profile.resticVersion; v == nil || v.LessThan(resticVersion14) {
		// restic < 0.14: from-repo => repo2, from-password-file => password-file2, etc.
		for name, legacyName := range legacyArgs {
			flags.Rename(name, legacyName)
		}
	}
	return
}

// BackupSection contains the specific configuration to the 'backup' command
type BackupSection struct {
	GenericSectionWithSchedule `mapstructure:",squash"`

	unresolvedSource  []string
	CheckBefore       bool     `mapstructure:"check-before" description:"Check the repository before starting the backup command"`
	CheckAfter        bool     `mapstructure:"check-after" description:"Check the repository after the backup command succeeded"`
	UseStdin          bool     `mapstructure:"stdin" argument:"stdin"`
	StdinCommand      []string `mapstructure:"stdin-command" description:"Shell command(s) that generate content to redirect into the stdin of restic. When set, the flag \"stdin\" is always set to \"true\"."`
	SourceRelative    bool     `mapstructure:"source-relative" description:"Enable backup with relative source paths. This will change the working directory of the \"restic backup\" command to \"source-base\", and will not expand \"source\" to an absolute path."`
	SourceBase        string   `mapstructure:"source-base" examples:"/;$PWD;C:\\;%cd%" description:"The base path to resolve relative backup paths against. Defaults to current directory if unset or empty (see also \"base-dir\" in profile)"`
	Source            []string `mapstructure:"source" examples:"/opt/;/home/user/;C:\\Users\\User\\Documents" description:"The paths to backup"`
	Exclude           []string `mapstructure:"exclude" argument:"exclude" argument-type:"no-glob"`
	Iexclude          []string `mapstructure:"iexclude" argument:"iexclude" argument-type:"no-glob"`
	ExcludeFile       []string `mapstructure:"exclude-file" argument:"exclude-file"`
	IexcludeFile      []string `mapstructure:"iexclude-file" argument:"iexclude-file"`
	FilesFrom         []string `mapstructure:"files-from" argument:"files-from"`
	FilesFromRaw      []string `mapstructure:"files-from-raw" argument:"files-from-raw"`
	FilesFromVerbatim []string `mapstructure:"files-from-verbatim" argument:"files-from-verbatim"`
	ExtendedStatus    bool     `mapstructure:"extended-status" argument:"json"`
	NoErrorOnWarning  bool     `mapstructure:"no-error-on-warning" description:"Do not fail the backup when some files could not be read"`
}

func (s *BackupSection) IsEmpty() bool { return s == nil }

func (b *BackupSection) resolve(profile *Profile) {
	b.ScheduleBaseSection.resolve(profile)

	// Ensure UseStdin is set when Backup.StdinCommand is defined
	if len(b.StdinCommand) > 0 {
		b.UseStdin = true
	}
	// Resolve symlinks if we send relative paths to restic (to match paths in snapshots)
	if b.SourceRelative {
		if dir := strings.TrimSpace(profile.BaseDir); dir != "" {
			profile.BaseDir = evaluateSymlinks(dir)
		}
		if dir := strings.TrimSpace(b.SourceBase); dir != "" {
			b.SourceBase = evaluateSymlinks(dir)
		}
	}
	// Resolve source paths
	if b.unresolvedSource == nil {
		b.unresolvedSource = b.Source
	}
	b.Source = profile.resolveSourcePath(b.SourceBase, b.SourceRelative, b.unresolvedSource...)

	// Extras, only enabled for Version >= 2 (to remain backward compatible in version 1)
	if profile.config != nil && profile.config.version >= Version02 {
		// Ensure that the host is in sync between backup & retention by setting it if missing
		if _, found := b.OtherFlags[constants.ParameterHost]; !found {
			b.SetOtherFlag(constants.ParameterHost, true)
		}
	}
}

func (s *BackupSection) setRootPath(p *Profile, rootPath string) {
	s.GenericSectionWithSchedule.setRootPath(p, rootPath)

	s.ExcludeFile = fixPaths(s.ExcludeFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	s.IexcludeFile = fixPaths(s.IexcludeFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	s.FilesFrom = fixPaths(s.FilesFrom, expandEnv, expandUserHome, absolutePrefix(rootPath))
	s.FilesFromRaw = fixPaths(s.FilesFromRaw, expandEnv, expandUserHome, absolutePrefix(rootPath))
	s.FilesFromVerbatim = fixPaths(s.FilesFromVerbatim, expandEnv, expandUserHome, absolutePrefix(rootPath))
	s.Exclude = fixPaths(s.Exclude, expandEnv, expandUserHome)
	s.Iexclude = fixPaths(s.Iexclude, expandEnv, expandUserHome)
}

// RetentionSection contains the specific configuration to
// the 'forget' command when running as part of a backup
type RetentionSection struct {
	ScheduleBaseSection `mapstructure:",squash" deprecated:"0.11.0"`
	OtherFlagsSection   `mapstructure:",squash"`

	BeforeBackup maybe.Bool `mapstructure:"before-backup" description:"Apply retention before starting the backup command"`
	AfterBackup  maybe.Bool `mapstructure:"after-backup" description:"Apply retention after the backup command succeeded. Defaults to true in configuration format v2 if any \"keep-*\" flag is set and \"before-backup\" is unset"`
}

func (r *RetentionSection) IsEmpty() bool { return r == nil }

func (r *RetentionSection) resolve(profile *Profile) {
	r.ScheduleBaseSection.resolve(profile)

	// Special cases of retention
	isSet := func(flags OtherFlags, name string) (found bool) { _, found = flags.GetOtherFlags()[name]; return }
	hasBackup := !profile.Backup.IsEmpty()

	// Copy "source" from "backup" as "path" if it hasn't been redefined
	if hasBackup && !isSet(r, constants.ParameterPath) {
		r.SetOtherFlag(constants.ParameterPath, true)
	}

	// Extras, only enabled for Version >= 2 (to remain backward compatible in version 1)
	if profile.config != nil && profile.config.version >= Version02 {
		// Auto-enable "after-backup" if nothing was specified explicitly and any "keep-" was configured
		if r.AfterBackup.IsUndefined() && r.BeforeBackup.IsUndefined() {
			for name := range r.OtherFlags {
				if strings.HasPrefix(name, "keep-") {
					r.AfterBackup = maybe.True()
					break
				}
			}
		}

		// Copy "tag" from "backup" if it was set and hasn't been redefined here
		// Allow setting it at profile level when not defined in "backup" nor "retention"
		if hasBackup &&
			!isSet(r, constants.ParameterTag) &&
			isSet(profile.Backup, constants.ParameterTag) {

			r.SetOtherFlag(constants.ParameterTag, true)
		}

		// Copy "host" from "backup" if it was set and hasn't been redefined here
		// Or use os.Hostname() same as restic does for backup when not setting it, see:
		// https://github.com/restic/restic/blob/master/cmd/restic/cmd_backup.go#L48
		if !isSet(r, constants.ParameterHost) {
			if hasBackup && isSet(profile.Backup, constants.ParameterHost) {
				r.SetOtherFlag(constants.ParameterHost, profile.Backup.OtherFlags[constants.ParameterHost])
			} else if !isSet(profile, constants.ParameterHost) {
				r.SetOtherFlag(constants.ParameterHost, true) // resolved with os.Hostname()
			}
		}
	}
}

// GenericSectionWithSchedule is a section containing schedule, shell command hooks and monitoring
// (all the other parameters being for restic)
type GenericSectionWithSchedule struct {
	GenericSection      `mapstructure:",squash"`
	ScheduleBaseSection `mapstructure:",squash"`
}

func (s *GenericSectionWithSchedule) setRootPath(p *Profile, rootPath string) {
	s.GenericSection.setRootPath(p, rootPath)
	s.ScheduleBaseSection.setRootPath(p, rootPath)
}

func (s *GenericSectionWithSchedule) IsEmpty() bool { return s == nil }

// ScheduleBaseSection contains the parameters for scheduling a command (backup, check, forget, etc.)
type ScheduleBaseSection struct {
	scheduleConfig                  *ScheduleConfig
	Schedule                        any            `mapstructure:"schedule" show:"noshow" examples:"hourly;daily;weekly;monthly;10:00,14:00,18:00,22:00;Wed,Fri 17:48;*-*-15 02:45;Mon..Fri 00:30" description:"Configures the scheduled execution of this profile section. Can be times in systemd timer format or a config structure"`
	SchedulePermission              string         `mapstructure:"schedule-permission" show:"noshow" default:"auto" enum:"auto;system;user;user_logged_on" description:"Specify whether the schedule runs with system or user privileges - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	ScheduleRunLevel                string         `mapstructure:"schedule-run-level" show:"noshow" default:"auto" enum:"auto;lowest;highest" description:"Specify the schedule privilege level (for Windows Task Scheduler only)"`
	ScheduleLog                     string         `mapstructure:"schedule-log" show:"noshow" examples:"/resticprofile.log;syslog-tcp://syslog-server:514;syslog:server;syslog:" description:"Redirect the output into a log file or to syslog when running on schedule"`
	SchedulePriority                string         `mapstructure:"schedule-priority" show:"noshow" default:"standard" enum:"background;standard" description:"Set the priority at which the schedule is run"`
	ScheduleLockMode                string         `mapstructure:"schedule-lock-mode" show:"noshow" default:"default" enum:"default;fail;ignore" description:"Specify how locks are used when running on schedule - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	ScheduleLockWait                maybe.Duration `mapstructure:"schedule-lock-wait" show:"noshow" examples:"150s;15m;30m;45m;1h;2h30m" description:"Set the maximum time to wait for acquiring locks when running on schedule"`
	ScheduleEnvCapture              []string       `mapstructure:"schedule-capture-environment" show:"noshow" default:"RESTIC_*" description:"Set names (or glob expressions) of environment variables to capture during schedule creation. The captured environment is applied prior to \"profile.env\" when running the schedule. Whether capturing is supported depends on the type of scheduler being used (supported in \"systemd\" and \"launchd\")"`
	ScheduleIgnoreOnBattery         maybe.Bool     `mapstructure:"schedule-ignore-on-battery" show:"noshow" default:"false" description:"Don't start this schedule when running on battery"`
	ScheduleIgnoreOnBatteryLessThan int            `mapstructure:"schedule-ignore-on-battery-less-than" show:"noshow" default:"" examples:"20;33;50;75" description:"Don't start this schedule when running on battery and the state of charge is less than this percentage"`
	ScheduleAfterNetworkOnline      maybe.Bool     `mapstructure:"schedule-after-network-online" show:"noshow" description:"Don't start this schedule when the network is offline (supported in \"systemd\")"`
	ScheduleHideWindow              maybe.Bool     `mapstructure:"schedule-hide-window" show:"noshow" default:"false" description:"Hide schedule window when running in foreground (Windows only)"`
	ScheduleStartWhenAvailable      maybe.Bool     `mapstructure:"schedule-start-when-available" show:"noshow" default:"false" description:"Start the task as soon as possible after a scheduled start is missed (Windows only)"`
}

func (s *ScheduleBaseSection) setRootPath(_ *Profile, _ string) {
	s.ScheduleLog = fixPath(s.ScheduleLog, expandEnv, expandUserHome)
}

func (s *ScheduleBaseSection) resolve(profile *Profile) {
	if s == nil || !profile.hasConfig() {
		return
	}
	if config := newScheduleConfig(profile, s); config.HasSchedules() {
		s.scheduleConfig = config
	}
}

func (s *ScheduleBaseSection) HasSchedule() bool { return s.scheduleConfig.HasSchedules() }

func (s *ScheduleBaseSection) getScheduleConfig(p *Profile, command string) *ScheduleConfig {
	if s.scheduleConfig != nil && p != nil {
		s.scheduleConfig.origin = ScheduleOrigin(p.Name, command)
	}
	return s.scheduleConfig
}

// CopySection contains the source or destination parameters for a copy command
type CopySection struct {
	GenericSectionWithSchedule `mapstructure:",squash"`

	Initialize                  bool              `mapstructure:"initialize" description:"Initialize the secondary repository if missing"`
	InitializeCopyChunkerParams maybe.Bool        `mapstructure:"initialize-copy-chunker-params" default:"true" description:"Copy chunker parameters when initializing the secondary repository"`
	FromRepository              ConfidentialValue `mapstructure:"from-repository" argument:"from-repo" description:"Source repository to copy snapshots from"`
	FromRepositoryFile          string            `mapstructure:"from-repository-file" argument:"from-repository-file" description:"File from which to read the source repository location to copy snapshots from"`
	FromPasswordFile            string            `mapstructure:"from-password-file" argument:"from-password-file" description:"File to read the source repository password from"`
	FromPasswordCommand         string            `mapstructure:"from-password-command" argument:"from-password-command" description:"Shell command to obtain the source repository password from"`
	FromKeyHint                 string            `mapstructure:"from-key-hint" argument:"from-key-hint" description:"Key ID of key to try decrypting the source repository first"`
	Snapshots                   []string          `mapstructure:"snapshot" description:"Snapshot IDs to copy (if empty, all snapshots are copied)"`
	ToRepository                ConfidentialValue `mapstructure:"repository" description:"Destination repository to copy snapshots to"`
	ToRepositoryFile            string            `mapstructure:"repository-file" description:"File from which to read the destination repository location to copy snapshots to"`
	ToPasswordFile              string            `mapstructure:"password-file" description:"File to read the destination repository password from"`
	ToPasswordCommand           string            `mapstructure:"password-command" description:"Shell command to obtain the destination repository password from"`
	ToKeyHint                   string            `mapstructure:"key-hint" description:"Key ID of key to try decrypting the destination repository first"`
}

func (s *CopySection) IsEmpty() bool { return s == nil }

func (s *CopySection) IsCopyTo() bool { return s.ToRepository.HasValue() || s.ToRepositoryFile != "" }

func (c *CopySection) resolve(p *Profile) {
	c.ScheduleBaseSection.resolve(p)

	if c.IsCopyTo() {
		c.ToRepository.setValue(fixPath(c.ToRepository.Value(), expandEnv, expandUserHome))
	} else {
		c.FromRepository.setValue(fixPath(c.FromRepository.Value(), expandEnv, expandUserHome))
	}
}

func (c *CopySection) setRootPath(p *Profile, rootPath string) {
	c.GenericSectionWithSchedule.setRootPath(p, rootPath)

	if c.IsCopyTo() {
		c.ToPasswordFile = fixPath(c.ToPasswordFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
		c.ToRepositoryFile = fixPath(c.ToRepositoryFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	} else {
		c.FromPasswordFile = fixPath(c.FromPasswordFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
		c.FromRepositoryFile = fixPath(c.FromRepositoryFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	}
}

func (s *CopySection) getInitFlags(profile *Profile) *shell.Args {
	var init *InitSection

	if s.IsCopyTo() {
		if s.InitializeCopyChunkerParams.IsTrueOrUndefined() {
			// Source repo for CopyChunkerParams
			init = &InitSection{
				CopyChunkerParams:   true,
				FromKeyHint:         profile.KeyHint,
				FromRepository:      profile.Repository,
				FromRepositoryFile:  profile.RepositoryFile,
				FromPasswordFile:    profile.PasswordFile,
				FromPasswordCommand: profile.PasswordCommand,
			}
			init.OtherFlags = profile.OtherFlags
		} else {
			init = new(InitSection)
		}

		// Repo that should be initialized
		ip := *profile
		ip.KeyHint = s.ToKeyHint
		ip.Repository = s.ToRepository
		ip.RepositoryFile = s.ToRepositoryFile
		ip.PasswordFile = s.ToPasswordFile
		ip.PasswordCommand = s.ToPasswordCommand
		ip.OtherFlags = s.OtherFlags
		return init.getCommandFlags(&ip)
	} else {
		if s.InitializeCopyChunkerParams.IsTrueOrUndefined() {
			// Source repo for CopyChunkerParams
			init = &InitSection{
				CopyChunkerParams:   true,
				FromKeyHint:         s.FromKeyHint,
				FromRepository:      s.FromRepository,
				FromRepositoryFile:  s.FromRepositoryFile,
				FromPasswordFile:    s.FromPasswordFile,
				FromPasswordCommand: s.FromPasswordCommand,
			}
			init.OtherFlags = profile.OtherFlags
		} else {
			init = new(InitSection)
		}

		// Repo that should be initialized
		return init.getCommandFlags(profile)
	}
}

func (s *CopySection) getCommandFlags(profile *Profile) (flags *shell.Args) {
	if s.IsCopyTo() {
		repositoryArgs := map[string]string{
			constants.ParameterRepository:      s.ToRepository.Value(),
			constants.ParameterRepositoryFile:  s.ToRepositoryFile,
			constants.ParameterPasswordFile:    s.ToPasswordFile,
			constants.ParameterPasswordCommand: s.ToPasswordCommand,
			constants.ParameterKeyHint:         s.ToKeyHint,
		}

		// Handle confidential repo in flags
		restore := profile.replaceWithRepositoryFile(&s.ToRepository, &s.ToRepositoryFile, "-to")
		defer restore()

		flags = profile.GetCommonFlags()
		addArgsFromStruct(flags, s)
		addArgsFromOtherFlags(flags, profile, s)

		if v := profile.resticVersion; v == nil || v.LessThan(resticVersion14) {
			// restic < 0.14: repo2, password-file2, etc. is the destination, repo, password-file, etc. the source
			for name, value := range repositoryArgs {
				if len(value) > 0 {
					flags.AddFlag(fmt.Sprintf("%s2", name), shell.NewArg(value, shell.ArgConfigEscape))
				}
			}
		} else {
			// restic >= 0.14: from-repo, from-password-file, etc. is the source, repo, password-file, etc. the destination
			for name := range maps.Keys(repositoryArgs) {
				flags.Rename(name, fmt.Sprintf("from-%s", name))
			}
			for name, value := range repositoryArgs {
				if len(value) > 0 {
					flags.AddFlag(name, shell.NewArg(value, shell.ArgConfigEscape))
				}
			}
		}
	} else {
		legacyArgs := map[string]string{
			"from-repo":             "repo2",
			"from-repository-file":  "repository-file2",
			"from-password-file":    "password-file2",
			"from-password-command": "password-command2",
			"from-key-hint":         "key-hint2",
		}

		// Handle confidential repo in flags
		restore := profile.replaceWithRepositoryFile(&s.FromRepository, &s.FromRepositoryFile, "-from")
		defer restore()

		flags = profile.GetCommonFlags()
		addArgsFromStruct(flags, s)
		addArgsFromOtherFlags(flags, profile, s)

		if v := profile.resticVersion; v == nil || v.LessThan(resticVersion14) {
			// restic < 0.14: from-repo => repo2, from-password-file => password-file2, etc.
			for name, legacyName := range legacyArgs {
				flags.Rename(name, legacyName)
			}
		}
	}

	// TODO: Handle positional option args (= shell.Args must support partitions)
	return
}

type StreamErrorSection struct {
	Pattern    string `mapstructure:"pattern" format:"regex" description:"A regular expression pattern that is tested against stderr of a running restic command"`
	MinMatches int    `mapstructure:"min-matches" range:"[0:]" description:"Minimum amount of times the \"pattern\" must match before \"run\" is started ; 0 for no limit"`
	MaxRuns    int    `mapstructure:"max-runs" range:"[0:]" description:"Maximum amount of times that \"run\" is started ; 0 for no limit"`
	Run        string `mapstructure:"run" description:"The shell command to run when the pattern matches"`
}

// RunShellCommandsSection is used to define shell commands that run before or after restic commands
type RunShellCommandsSection struct {
	RunBefore    []string `mapstructure:"run-before" description:"Run shell command(s) before a restic command"`
	RunAfter     []string `mapstructure:"run-after" description:"Run shell command(s) after a successful restic command"`
	RunAfterFail []string `mapstructure:"run-after-fail" description:"Run shell command(s) after failed restic or shell commands"`
	RunFinally   []string `mapstructure:"run-finally" description:"Run shell command(s) always, after all other commands"`
}

func (r *RunShellCommandsSection) GetRunShellCommands() *RunShellCommandsSection { return r }

// OtherFlagsSection contains additional restic command line flags
type OtherFlagsSection struct {
	OtherFlags map[string]any `mapstructure:",remain"`
}

func (o *OtherFlagsSection) GetOtherFlags() map[string]any { return o.OtherFlags }

func (o *OtherFlagsSection) SetOtherFlag(name string, value any) {
	if o.OtherFlags == nil {
		o.OtherFlags = make(map[string]any)
	}
	o.OtherFlags[name] = value
}

// NewProfile instantiates a new blank profile
func NewProfile(c *Config, name string) (p *Profile) {
	p = &Profile{
		Name:                 name,
		config:               c,
		OtherSections:        make(map[string]*GenericSection),
		PrometheusPushFormat: constants.DefaultPrometheusPushFormat,
	}

	// create dynamic sections defined in any known restic version
	sectionStructs := p.allSectionStructs()
	for _, command := range restic.CommandNamesForVersion(restic.AnyVersion) {
		if _, hasSectionStruct := sectionStructs[command]; hasSectionStruct {
			continue
		}
		p.OtherSections[command] = nil // set the key only, section remains empty by default
	}
	return
}

// fillOtherSections transfers parsed configuration from OtherFlags into OtherSections
func (p *Profile) fillOtherSections() {
	if p.OtherFlags == nil {
		return
	}

	for name := range p.OtherSections {
		if content := p.OtherFlags[name]; content != nil {
			section := new(GenericSection)

			var err error
			if !p.hasConfig() {
				err = mapstructure.WeakDecode(content, section)
			} else {
				if decoder, e := p.config.newUnmarshaller(section); e == nil {
					err = decoder.Decode(content)
				} else {
					err = e
				}
			}

			if err == nil {
				p.OtherSections[name] = section
				delete(p.OtherFlags, name)
			} else if p.hasConfig() {
				p.config.reportFailedSection(name, err)
			}
		}
	}
}

// ResolveConfiguration resolves dependencies between profile config flags
func (p *Profile) ResolveConfiguration() {
	p.fillOtherSections()

	// Resolve paths that do not depend on root path
	p.BaseDir = fixPath(p.BaseDir, expandEnv, expandUserHome)
	p.Repository.setValue(fixPath(p.Repository.Value(), expandEnv, expandUserHome))

	// Resolve all sections implementing resolver
	for _, r := range GetSectionsWith[resolver](p) {
		r.resolve(p)
	}

	// Resolve environment variable name case (p.Environment keys are all lower case due to config parser)
	// Custom env variables (without a match in os.Environ) are changed to uppercase (like before in wrapper)
	osEnv := util.NewFoldingEnvironment(os.Environ()...)
	for name, value := range p.Environment {
		if newName := osEnv.ResolveName(strings.ToUpper(name)); newName != name {
			delete(p.Environment, name)
			p.Environment[newName] = value
		}
	}

	// Deal with "path" & "tag" flags
	if p.Backup != nil {
		// Copy tags from backup if tag is set to boolean true
		if tags, ok := stringifyValueOf(p.Backup.OtherFlags[constants.ParameterTag]); ok {
			p.SetTag(strings.Join(tags, ",")) // must use "tag1,tag2,..." to require all tags
		} else {
			p.SetTag() // resolve tag parameters when no tag is set in backup
		}

		// Copy parameter path from backup sources if path is set to boolean true
		p.SetPath(p.Backup.SourceBase, p.Backup.unresolvedSource...)
	} else {
		// Resolve path parameter (no copy since backup is not defined)
		p.SetPath("")
	}
}

// SetResticVersion sets the effective restic version for validation and to determine how to format flags.
// Note that flags filtering happens later inside resticWrapper and is not necessary inside the profile.
func (p *Profile) SetResticVersion(resticVersion string) (err error) {
	if len(resticVersion) == 0 {
		p.resticVersion = nil
	} else {
		p.resticVersion, err = semver.NewVersion(resticVersion)
	}
	return
}

// SetRootPath changes the path of all the relative paths and files in the configuration
func (p *Profile) SetRootPath(rootPath string) {
	p.Lock = fixPath(p.Lock, expandEnv, absolutePrefix(rootPath))
	p.PasswordFile = fixPath(p.PasswordFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	p.RepositoryFile = fixPath(p.RepositoryFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	p.CacheDir = fixPath(p.CacheDir, expandEnv, expandUserHome, absolutePrefix(rootPath))
	p.CACert = fixPath(p.CACert, expandEnv, expandUserHome, absolutePrefix(rootPath))
	p.TLSClientCert = fixPath(p.TLSClientCert, expandEnv, expandUserHome, absolutePrefix(rootPath))
	p.SystemdDropInFiles = fixPaths(p.SystemdDropInFiles, expandEnv, absolutePrefix(rootPath))
	p.EnvironmentFiles = fixPaths(p.EnvironmentFiles, expandEnv, expandUserHome, absolutePrefix(rootPath))

	// Forward to sections accepting paths
	for _, s := range GetSectionsWith[relativePath](p) {
		s.setRootPath(p, rootPath)
	}

	// Handle dynamic flags dealing with paths that are relative to root path
	filepathFlags := []string{
		"cacert",
		"tls-client-cert",
		"cache-dir",
		constants.ParameterPasswordFile,
		constants.ParameterRepositoryFile,
	}
	for _, section := range p.allFlagsSections() {
		for _, flag := range filepathFlags {
			if paths, ok := stringifyValueOf(section[flag]); ok && len(paths) > 0 {
				for i, path := range paths {
					if len(path) > 0 {
						paths[i] = fixPath(path, expandEnv, expandUserHome, absolutePrefix(rootPath))
					}
				}
				section[flag] = paths
			}
		}
	}
}

func (p *Profile) resolveSourcePath(sourceBase string, relativePaths bool, sourcePaths ...string) []string {
	var applySourceBase, applyBaseDir pathFix

	sourceBase = fixPath(strings.TrimSpace(sourceBase), expandEnv, expandUserHome)
	// When "source-relative" is set, the source paths are relative to the "source-base"
	if !relativePaths {
		// Backup source is NOT relative to the configuration, but to PWD or sourceBase (if not empty)
		// Applying "sourceBase" if set
		if sourceBase != "" {
			applySourceBase = absolutePrefix(sourceBase)
		}
		// Applying a custom PWD eagerly so that own commands (e.g. "show") display correct paths
		if p.BaseDir != "" {
			applyBaseDir = absolutePrefix(p.BaseDir)
		}
	} else if p.BaseDir == "" && sourceBase == "" && p.hasConfig() {
		p.config.reportChangedPath(".", "<none>", "source-base (for relative source)")
	}

	// prefix paths starting with "-" with a "./" to distinguish a source path from a flag
	maskPathsWithFlagPrefix := func(file string) string {
		if strings.HasPrefix(file, "-") {
			return "." + string(filepath.Separator) + file
		}
		return file
	}

	sourcePaths = fixPaths(sourcePaths, expandEnv, expandUserHome, applySourceBase, applyBaseDir)
	sourcePaths = resolveGlob(sourcePaths)
	sourcePaths = fixPaths(sourcePaths, maskPathsWithFlagPrefix, filepath.ToSlash, filepath.FromSlash)
	return sourcePaths
}

// SetHost will replace any host value from a boolean to the hostname
func (p *Profile) SetHost(hostname string) {
	for _, section := range p.allFlagsSections() {
		replaceTrueValue(section, constants.ParameterHost, hostname)
	}
}

// SetTag will replace any tag value from a boolean to the tags
func (p *Profile) SetTag(tags ...string) {
	for _, section := range p.allFlagsSections() {
		replaceTrueValue(section, constants.ParameterTag, tags...)
	}
}

// SetPath will replace any path value from a boolean to sourcePaths and change paths to absolute
func (p *Profile) SetPath(basePath string, sourcePaths ...string) {
	hasAbsoluteBase := filepath.IsAbs(p.BaseDir) || filepath.IsAbs(basePath)

	resolvePath := func(origin string, paths []string, revolver func(string) []string) (resolved []string) {
		for _, path := range paths {
			if len(path) > 0 {
				for _, rp := range revolver(path) {
					if rp != path && p.hasConfig() && !hasAbsoluteBase {
						p.config.reportChangedPath(rp, path, origin)
					}
					resolved = append(resolved, rp)
				}
			}
		}
		return resolved
	}

	sourcePathsResolved := false

	// Resolve 'path' to absolute paths as anything else will not select any snapshots
	for _, section := range p.allFlagsSections() {
		value, hasValue := section[constants.ParameterPath]
		if !hasValue {
			continue
		}

		if replace, ok := value.(bool); ok && replace {
			// Replace bool-true with absolute sourcePaths
			if !sourcePathsResolved {
				sourcePaths = resolvePath("path (from source)", sourcePaths, func(path string) []string {
					return fixPaths(p.resolveSourcePath(basePath, false, path), absolutePath)
				})
				sourcePathsResolved = true
			}
			section[constants.ParameterPath] = sourcePaths

		} else if paths, ok := stringifyValueOf(value); ok && len(paths) > 0 {
			// Resolve path strings to absolute paths
			paths = resolvePath("path", paths, func(path string) []string {
				return []string{fixPath(path, expandEnv, absolutePath)}
			})
			section[constants.ParameterPath] = paths
		}
	}
}

func (p *Profile) allFlagsSections() (sections []map[string]any) {
	for _, section := range GetSectionsWith[OtherFlags](p) {
		if flags := section.GetOtherFlags(); flags != nil {
			sections = append(sections, flags)
		}
	}
	return
}

func (p *Profile) replaceWithRepositoryFile(repository *ConfidentialValue, repositoryFile *string, suffix string) (restore func()) {
	origRepo, origFile := *repository, *repositoryFile
	restore = func() {
		*repository = origRepo
		*repositoryFile = origFile
	}

	// abort if repo is not confidential or a file is in use already
	if !repository.IsConfidential() || len(origFile) > 0 {
		return
	}

	// abort by global config or environment
	var global *Global
	if p.hasConfig() {
		global = p.config.mustGetGlobalSection()
	}
	if global == nil || global.NoAutoRepositoryFile.IsTrue() {
		return
	} else if global.NoAutoRepositoryFile.IsUndefined() {
		env := p.GetEnvironment(true)
		if len(env.Get("RESTIC_REPOSITORY")) > 0 || len(env.Get("RESTIC_REPOSITORY_FILE")) > 0 {
			clog.Debug("restic repository is set using environment variables, not replacing plain \"repository\" argument")
			return
		}
	}

	// apply temporary change
	file, err := templates.PrivateTempFile(fmt.Sprintf("%s%s-repo.txt", p.Name, suffix))
	if err != nil {
		clog.Debugf(`private file %s not supported: %s`, file, err.Error())
		return
	}

	if err = os.WriteFile(file, []byte(origRepo.Value()), 0600); err == nil {
		clog.Debugf(`replaced plain "repository" argument with "repository-file" (%s) to avoid password leak`, file)
		*repository = NewConfidentialValue("")
		*repositoryFile = file
	} else {
		clog.Debugf(`failed writing %s: %s`, file, err.Error())
	}
	return
}

// hasConfig returns true if the profile has an initialized config instance (always true in normal runtime)
func (p *Profile) hasConfig() bool {
	return p != nil &&
		p.config != nil &&
		p.config.viper != nil
}

// GetCommonFlags returns the flags common to all commands
func (p *Profile) GetCommonFlags() (flags *shell.Args) {
	// Handle confidential repo in flags
	restore := p.replaceWithRepositoryFile(&p.Repository, &p.RepositoryFile, "")
	defer restore()

	// Flags from the profile fields
	flags = shell.NewArgs()
	addArgsFromStruct(flags, p)
	addArgsFromOtherFlags(flags, p, p)
	return flags
}

// GetCommandFlags returns the flags specific to the command (backup, snapshots, forget, etc.)
func (p *Profile) GetCommandFlags(command string) (flags *shell.Args) {
	if section, ok := GetSectionWith[commandFlags](p, command); ok {
		// Section specific implementation
		flags = section.getCommandFlags(p)
	} else {
		// Default implementation
		flags = p.GetCommonFlags()
		if section, ok := GetSectionWith[any](p, command); ok {
			addArgsFromStruct(flags, section)
		}
		if section, ok := GetSectionWith[OtherFlags](p, command); ok {
			addArgsFromOtherFlags(flags, p, section)
		}
	}
	return flags
}

// GetCopyInitializeFlags returns the flags specific to the "init" command when used to initialize the copy destination
func (p *Profile) GetCopyInitializeFlags() (args *shell.Args) {
	if p.Copy != nil {
		args = p.Copy.getInitFlags(p)
	}
	return
}

// GetRetentionFlags returns the flags specific to the "forget" command being run as part of a backup
func (p *Profile) GetRetentionFlags() *shell.Args {
	return p.GetCommandFlags(constants.SectionConfigurationRetention)
}

// HasDeprecatedRetentionSchedule indicates if there's one or more schedule parameters in the retention section,
// which is deprecated as of 0.11.0
func (p *Profile) HasDeprecatedRetentionSchedule() bool {
	return p.Retention != nil && p.Retention.HasSchedule()
}

// GetBackupSource returns the directories to back up
func (p *Profile) GetBackupSource() []string {
	if p.Backup == nil {
		return nil
	}
	return p.Backup.Source
}

// GetCopySnapshotIDs returns the snapshot IDs to copy
func (p *Profile) GetCopySnapshotIDs() []string {
	if p.Copy == nil {
		return nil
	}
	return p.Copy.Snapshots
}

// DefinedCommands returns all commands (also called sections) defined in the profile (backup, check, forget, etc.)
func (p *Profile) DefinedCommands() []string {
	return slices.Sorted(maps.Keys(GetSectionsWith[any](p)))
}

func (p *Profile) allSectionStructs() map[string]any {
	return map[string]any{
		constants.CommandBackup:                 p.Backup,
		constants.CommandCheck:                  p.Check,
		constants.CommandCopy:                   p.Copy,
		constants.CommandForget:                 p.Forget,
		constants.CommandPrune:                  p.Prune,
		constants.CommandInit:                   p.Init,
		constants.SectionConfigurationRetention: p.Retention,
	}
}

// AllSections returns all possible sections of this profile (including undefined sections set to nil)
func (p *Profile) AllSections() (sections map[string]any) {
	sections = p.allSectionStructs()
	for name, section := range p.OtherSections {
		sections[name] = section
	}
	return
}

// SchedulableCommands returns all command names that could have a schedule
func (p *Profile) SchedulableCommands() []string {
	return slices.Sorted(maps.Keys(GetDeclaredSectionsWith[scheduling](p)))
}

func (p *Profile) GetEnvironment(withOs bool) (env *util.Environment) {
	env = util.NewDefaultEnvironment()

	// OS environment
	if withOs {
		env.SetValues(os.Environ()...)
	}

	// Profile environment files
	for _, file := range p.EnvironmentFiles {
		if ef, err := util.GetEnvironmentFile(file); err == nil {
			ef.AddTo(env)
		} else {
			clog.Debugf("failed loading dotenv %s: %s", file, err.Error())
		}
	}

	// Profile environment
	for key, value := range p.Environment {
		env.Put(key, value.Value())
	}

	return
}

// Schedules returns a map of command -> Schedule, for all the commands that have a schedule configuration
func (p *Profile) Schedules() map[string]*Schedule {
	// All SectionWithSchedule (backup, check, prune, etc.)
	sections := GetSectionsWith[scheduling](p)
	schedules := make(map[string]*Schedule)

	for name, section := range sections {
		if config := section.getScheduleConfig(p, name); config != nil {
			schedules[name] = newScheduleForProfile(p, config)
		}
	}

	return schedules
}

func (p *Profile) GetRunShellCommandsSections(command string) (profileCommands RunShellCommandsSection, sectionCommands RunShellCommandsSection) {
	if c := p.GetRunShellCommands(); c != nil {
		profileCommands = *c
	}

	if section, ok := GetSectionWith[RunShellCommands](p, command); ok {
		if c := section.GetRunShellCommands(); c != nil {
			sectionCommands = *c
		}
	}
	return
}

func (p *Profile) GetMonitoringSections(command string) (monitoring SendMonitoringSections) {
	if section, ok := GetSectionWith[Monitoring](p, command); ok {
		monitoring = *section.GetSendMonitoring()
	}
	return
}

func (o *Profile) Kind() string {
	return constants.SchedulableKindProfile
}

// GetDeclaredSectionsWith returns all sections that implement a certain interface (including nil values)
func GetDeclaredSectionsWith[T any](p *Profile) (sections map[string]T) {
	sections = make(map[string]T)
	for name, section := range p.AllSections() {
		if targetType, canCast := section.(T); canCast {
			sections[name] = targetType
		}
	}
	return
}

// GetSectionsWith returns all sections that implement a certain interface (excluding nil values)
func GetSectionsWith[T any](p *Profile) (sections map[string]T) {
	sections = GetDeclaredSectionsWith[T](p)
	for name, section := range sections {
		if isEmpty(section) {
			delete(sections, name)
		}
	}
	return
}

// GetSectionWith returns a section that implement a certain interface (excluding nil values)
func GetSectionWith[T any](p *Profile, name string) (result T, ok bool) {
	if result, ok = GetDeclaredSectionsWith[T](p)[name]; ok {
		ok = !isEmpty(result)
	}
	return
}

func isEmpty(section any) bool {
	if e, ok := section.(Empty); ok {
		return e.IsEmpty()
	}
	return reflect.ValueOf(section).IsNil()
}

func replaceTrueValue(source map[string]any, key string, replace ...string) {
	if genericValue, ok := source[key]; ok {
		if value, ok := genericValue.(bool); ok {
			if value {
				if len(replace) > 0 {
					source[key] = replace
				} else {
					delete(source, key)
				}
			}
		}
	}
}

func addArgsFromOtherFlags(args *shell.Args, profile *Profile, section OtherFlags) {
	aliases := argAliasesFromStruct(profile)
	maps.Copy(aliases, argAliasesFromStruct(section))
	addArgsFromMap(args, aliases, section.GetOtherFlags())
}

// Implements Schedulable
var _ Schedulable = new(Profile)
