package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/bools"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// resticVersion14 is the semver of restic 0.14 (the version where several flag names were changed)
var resticVersion14 = semver.MustParse("0.14")

// Empty allows to test if a section is specified or not
type Empty interface {
	IsEmpty() bool
}

// Scheduling provides access to schedule information inside a section
type Scheduling interface {
	GetSchedule() *ScheduleBaseSection
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
	config                  *Config
	legacyArg               bool
	resticVersion           *semver.Version
	Name                    string
	Description             string                            `mapstructure:"description" description:"Describes the profile"`
	BaseDir                 string                            `mapstructure:"base-dir" description:"Sets the working directory for this profile. The profile will fail when the working directory cannot be changed. Leave empty to use the current directory instead"`
	Quiet                   bool                              `mapstructure:"quiet" argument:"quiet"`
	Verbose                 int                               `mapstructure:"verbose" argument:"verbose"`
	KeyHint                 string                            `mapstructure:"key-hint" argument:"key-hint"`
	Repository              ConfidentialValue                 `mapstructure:"repository" argument:"repo"`
	RepositoryFile          string                            `mapstructure:"repository-file" argument:"repository-file"`
	PasswordFile            string                            `mapstructure:"password-file" argument:"password-file"`
	PasswordCommand         string                            `mapstructure:"password-command" argument:"password-command"`
	CacheDir                string                            `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert                  string                            `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert           string                            `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize              bool                              `mapstructure:"initialize" default:"" description:"Initialize the restic repository if missing"`
	Inherit                 string                            `mapstructure:"inherit" show:"noshow" description:"Name of the profile to inherit all of the settings from"`
	Lock                    string                            `mapstructure:"lock" description:"Path to the lock file to use with resticprofile locks"`
	ForceLock               bool                              `mapstructure:"force-inactive-lock" description:"Allows to lock when the existing lock is considered stale"`
	StreamError             []StreamErrorSection              `mapstructure:"stream-error" description:"Run shell command(s) when a pattern matches the stderr of restic"`
	StatusFile              string                            `mapstructure:"status-file" description:"Path to the status file to update with a summary of last restic command result"`
	PrometheusSaveToFile    string                            `mapstructure:"prometheus-save-to-file" description:"Path to the prometheus metrics file to update with a summary of the last restic command result"`
	PrometheusPush          string                            `mapstructure:"prometheus-push" format:"uri" description:"URL of the prometheus push gateway to send the summary of the last restic command result to"`
	PrometheusPushJob       string                            `mapstructure:"prometheus-push-job" description:"Prometheus push gateway job name. $command placeholder is replaced with restic command"`
	PrometheusLabels        map[string]string                 `mapstructure:"prometheus-labels" description:"Additional prometheus labels to set"`
	Environment             map[string]ConfidentialValue      `mapstructure:"env" description:"Additional environment variables to set in any child process"`
	Init                    *InitSection                      `mapstructure:"init"`
	Backup                  *BackupSection                    `mapstructure:"backup"`
	Retention               *RetentionSection                 `mapstructure:"retention" command:"forget"`
	Check                   *SectionWithScheduleAndMonitoring `mapstructure:"check"`
	Prune                   *SectionWithScheduleAndMonitoring `mapstructure:"prune"`
	Forget                  *SectionWithScheduleAndMonitoring `mapstructure:"forget"`
	Copy                    *CopySection                      `mapstructure:"copy"`
	OtherSections           map[string]*GenericSection        `show:",remain"`
}

// GenericSection is used for all restic commands that are not covered in specific section types
type GenericSection struct {
	OtherFlagsSection       `mapstructure:",squash"`
	RunShellCommandsSection `mapstructure:",squash"`
}

func (g *GenericSection) IsEmpty() bool { return g == nil }

// InitSection contains the specific configuration to the 'init' command
type InitSection struct {
	OtherFlagsSection   `mapstructure:",squash"`
	CopyChunkerParams   bool              `mapstructure:"copy-chunker-params" argument:"copy-chunker-params"`
	FromKeyHint         string            `mapstructure:"from-key-hint" argument:"from-key-hint"`
	FromRepository      ConfidentialValue `mapstructure:"from-repository" argument:"from-repo"`
	FromRepositoryFile  string            `mapstructure:"from-repository-file" argument:"from-repository-file"`
	FromPasswordFile    string            `mapstructure:"from-password-file" argument:"from-password-file"`
	FromPasswordCommand string            `mapstructure:"from-password-command" argument:"from-password-command"`
}

func (i *InitSection) IsEmpty() bool { return i == nil }

func (i *InitSection) resolve(p *Profile) {
	i.FromRepository.setValue(fixPath(i.FromRepository.Value(), expandEnv, expandUserHome))
}

func (i *InitSection) setRootPath(_ *Profile, rootPath string) {
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
	SectionWithScheduleAndMonitoring `mapstructure:",squash"`
	RunShellCommandsSection          `mapstructure:",squash"`
	unresolvedSource                 []string
	CheckBefore                      bool     `mapstructure:"check-before" description:"Check the repository before starting the backup command"`
	CheckAfter                       bool     `mapstructure:"check-after" description:"Check the repository after the backup command succeeded"`
	UseStdin                         bool     `mapstructure:"stdin" argument:"stdin"`
	StdinCommand                     []string `mapstructure:"stdin-command" description:"Shell command(s) that generate content to redirect into the stdin of restic. When set, the flag \"stdin\" is always set to \"true\"."`
	SourceBase                       string   `mapstructure:"source-base" examples:"/;$PWD;C:\\;%cd%" description:"The base path to resolve relative backup paths against. Defaults to current directory if unset or empty (see also \"base-dir\" in profile)"`
	Source                           []string `mapstructure:"source" examples:"/opt/;/home/user/;C:\\Users\\User\\Documents" description:"The paths to backup"`
	Exclude                          []string `mapstructure:"exclude" argument:"exclude" argument-type:"no-glob"`
	Iexclude                         []string `mapstructure:"iexclude" argument:"iexclude" argument-type:"no-glob"`
	ExcludeFile                      []string `mapstructure:"exclude-file" argument:"exclude-file"`
	IexcludeFile                     []string `mapstructure:"iexclude-file" argument:"iexclude-file"`
	FilesFrom                        []string `mapstructure:"files-from" argument:"files-from"`
	FilesFromRaw                     []string `mapstructure:"files-from-raw" argument:"files-from-raw"`
	FilesFromVerbatim                []string `mapstructure:"files-from-verbatim" argument:"files-from-verbatim"`
	ExtendedStatus                   bool     `mapstructure:"extended-status" argument:"json"`
	NoErrorOnWarning                 bool     `mapstructure:"no-error-on-warning" description:"Do not fail the backup when some files could not be read"`
}

func (s *BackupSection) IsEmpty() bool { return s == nil }

func (b *BackupSection) resolve(profile *Profile) {
	// Ensure UseStdin is set when Backup.StdinCommand is defined
	if len(b.StdinCommand) > 0 {
		b.UseStdin = true
	}
	// Resolve source paths
	if b.unresolvedSource == nil {
		b.unresolvedSource = b.Source
	}
	b.Source = profile.resolveSourcePath(b.SourceBase, b.unresolvedSource...)

	// Extras, only enabled for Version >= 2 (to remain backward compatible in version 1)
	if profile.config != nil && profile.config.version >= Version02 {
		// Ensure that the host is in sync between backup & retention by setting it if missing
		if _, found := b.OtherFlags[constants.ParameterHost]; !found {
			b.SetOtherFlag(constants.ParameterHost, true)
		}
	}
}

func (s *BackupSection) setRootPath(p *Profile, rootPath string) {
	s.SectionWithScheduleAndMonitoring.setRootPath(p, rootPath)

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
	BeforeBackup        *bool `mapstructure:"before-backup" description:"Apply retention before starting the backup command"`
	AfterBackup         *bool `mapstructure:"after-backup" description:"Apply retention after the backup command succeeded. Defaults to true if any \"keep-*\" flag is set and \"before-backup\" is unset"`
}

func (r *RetentionSection) IsEmpty() bool { return r == nil }

func (r *RetentionSection) resolve(profile *Profile) {
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
		if bools.IsUndefined(r.AfterBackup) && bools.IsUndefined(r.BeforeBackup) {
			for name, _ := range r.OtherFlags {
				if strings.HasPrefix(name, "keep-") {
					r.AfterBackup = bools.True()
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

// SectionWithScheduleAndMonitoring is a section containing schedule, shell command hooks and monitoring
// (all the other parameters being for restic)
type SectionWithScheduleAndMonitoring struct {
	ScheduleBaseSection    `mapstructure:",squash"`
	SendMonitoringSections `mapstructure:",squash"`
	OtherFlagsSection      `mapstructure:",squash"`
}

func (s *SectionWithScheduleAndMonitoring) setRootPath(p *Profile, rootPath string) {
	s.SendMonitoringSections.setRootPath(p, rootPath)
	s.ScheduleBaseSection.setRootPath(p, rootPath)
}

func (s *SectionWithScheduleAndMonitoring) IsEmpty() bool { return s == nil }

// ScheduleBaseSection contains the parameters for scheduling a command (backup, check, forget, etc.)
type ScheduleBaseSection struct {
	Schedule           []string      `mapstructure:"schedule" show:"noshow" examples:"hourly;daily;weekly;monthly;10:00,14:00,18:00,22:00;Wed,Fri 17:48;*-*-15 02:45;Mon..Fri 00:30" description:"Set the times at which the scheduled command is run (times are specified in systemd timer format)"`
	SchedulePermission string        `mapstructure:"schedule-permission" show:"noshow" default:"auto" enum:"auto;system;user;user_logged_on" description:"Specify whether the schedule runs with system or user privileges - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	ScheduleLog        string        `mapstructure:"schedule-log" show:"noshow" examples:"/resticprofile.log;tcp://localhost:514" description:"Redirect the output into a log file or to syslog when running on schedule"`
	SchedulePriority   string        `mapstructure:"schedule-priority" show:"noshow" default:"background" enum:"background;standard" description:"Set the priority at which the schedule is run"`
	ScheduleLockMode   string        `mapstructure:"schedule-lock-mode" show:"noshow" default:"default" enum:"default;fail;ignore" description:"Specify how locks are used when running on schedule - see https://creativeprojects.github.io/resticprofile/schedules/configuration/"`
	ScheduleLockWait   time.Duration `mapstructure:"schedule-lock-wait" show:"noshow" examples:"150s;15m;30m;45m;1h;2h30m" description:"Set the maximum time to wait for acquiring locks when running on schedule"`
	ScheduleEnvCapture []string      `mapstructure:"schedule-capture-environment" show:"noshow" default:"RESTIC_*" description:"Set names (or glob expressions) of environment variables to capture during schedule creation. The captured environment is applied prior to \"profile.env\" when running the schedule. Whether capturing is supported depends on the type of scheduler being used (supported in \"systemd\" and \"launchd\")"`
}

func (s *ScheduleBaseSection) setRootPath(_ *Profile, _ string) {
	s.ScheduleLog = fixPath(s.ScheduleLog, expandEnv, expandUserHome)
}

func (s *ScheduleBaseSection) GetSchedule() *ScheduleBaseSection {
	if s != nil && s.ScheduleEnvCapture == nil {
		s.ScheduleEnvCapture = []string{"RESTIC_*"}
	}
	return s
}

// CopySection contains the destination parameters for a copy command
type CopySection struct {
	SectionWithScheduleAndMonitoring `mapstructure:",squash"`
	RunShellCommandsSection          `mapstructure:",squash"`
	Initialize                       bool              `mapstructure:"initialize" description:"Initialize the secondary repository if missing"`
	InitializeCopyChunkerParams      *bool             `mapstructure:"initialize-copy-chunker-params" default:"true" description:"Copy chunker parameters when initializing the secondary repository"`
	Repository                       ConfidentialValue `mapstructure:"repository" description:"Destination repository to copy snapshots to"`
	RepositoryFile                   string            `mapstructure:"repository-file" description:"File from which to read the destination repository location to copy snapshots to"`
	PasswordFile                     string            `mapstructure:"password-file" description:"File to read the destination repository password from"`
	PasswordCommand                  string            `mapstructure:"password-command" description:"Shell command to obtain the destination repository password from"`
	KeyHint                          string            `mapstructure:"key-hint" description:"Key ID of key to try decrypting the destination repository first"`
}

func (s *CopySection) IsEmpty() bool { return s == nil }

func (c *CopySection) resolve(p *Profile) {
	c.Repository.setValue(fixPath(c.Repository.Value(), expandEnv, expandUserHome))
}

func (c *CopySection) setRootPath(p *Profile, rootPath string) {
	c.SectionWithScheduleAndMonitoring.setRootPath(p, rootPath)

	c.PasswordFile = fixPath(c.PasswordFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
	c.RepositoryFile = fixPath(c.RepositoryFile, expandEnv, expandUserHome, absolutePrefix(rootPath))
}

func (s *CopySection) getInitFlags(profile *Profile) *shell.Args {
	var init *InitSection

	if bools.IsTrueOrUndefined(s.InitializeCopyChunkerParams) {
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
	ip.KeyHint = s.KeyHint
	ip.Repository = s.Repository
	ip.RepositoryFile = s.RepositoryFile
	ip.PasswordFile = s.PasswordFile
	ip.PasswordCommand = s.PasswordCommand
	ip.OtherFlags = s.OtherFlags

	return init.getCommandFlags(&ip)
}

func (s *CopySection) getCommandFlags(profile *Profile) (flags *shell.Args) {
	repositoryArgs := map[string]string{
		constants.ParameterRepository:      s.Repository.Value(),
		constants.ParameterRepositoryFile:  s.RepositoryFile,
		constants.ParameterPasswordFile:    s.PasswordFile,
		constants.ParameterPasswordCommand: s.PasswordCommand,
		constants.ParameterKeyHint:         s.KeyHint,
	}

	flags = profile.GetCommonFlags()
	addArgsFromStruct(flags, s)
	addArgsFromOtherFlags(flags, profile, s)

	if v := profile.resticVersion; v == nil || v.LessThan(resticVersion14) {
		// restic < 0.14: repo2, password-file2, etc. is the destination, repo, password-file, etc. the source
		for name, value := range repositoryArgs {
			if len(value) > 0 {
				flags.AddFlag(fmt.Sprintf("%s2", name), value, shell.ArgConfigEscape)
			}
		}
	} else {
		// restic >= 0.14: from-repo, from-password-file, etc. is the source, repo, password-file, etc. the destination
		for _, name := range maps.Keys(repositoryArgs) {
			flags.Rename(name, fmt.Sprintf("from-%s", name))
		}
		for name, value := range repositoryArgs {
			if len(value) > 0 {
				flags.AddFlag(name, value, shell.ArgConfigEscape)
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

// SendMonitoringSections is a group of target to send monitoring information
type SendMonitoringSections struct {
	SendBefore    []SendMonitoringSection `mapstructure:"send-before" description:"Send HTTP request(s) before a restic command"`
	SendAfter     []SendMonitoringSection `mapstructure:"send-after" description:"Send HTTP request(s) after a successful restic command"`
	SendAfterFail []SendMonitoringSection `mapstructure:"send-after-fail" description:"Send HTTP request(s) after failed restic or shell commands"`
	SendFinally   []SendMonitoringSection `mapstructure:"send-finally" description:"Send HTTP request(s) always, after all other commands"`
}

func (s *SendMonitoringSections) setRootPath(_ *Profile, rootPath string) {
	for _, monitoringSections := range s.getAllSendMonitoringSections() {
		for index, value := range monitoringSections {
			monitoringSections[index].BodyTemplate = fixPath(value.BodyTemplate, expandEnv, expandUserHome, absolutePrefix(rootPath))
		}
	}
}

func (s *SendMonitoringSections) GetSendMonitoring() *SendMonitoringSections { return s }

func (s *SendMonitoringSections) getAllSendMonitoringSections() [][]SendMonitoringSection {
	return [][]SendMonitoringSection{
		s.SendBefore,
		s.SendAfter,
		s.SendAfterFail,
		s.SendFinally,
	}
}

// SendMonitoringSection is used to send monitoring information to third party software
type SendMonitoringSection struct {
	Method       string                 `mapstructure:"method" enum:"GET;DELETE;HEAD;OPTIONS;PATCH;POST;PUT;TRACE" default:"GET" description:"HTTP method of the request"`
	URL          ConfidentialValue      `mapstructure:"url" format:"uri" description:"URL of the target to send to"`
	Headers      []SendMonitoringHeader `mapstructure:"headers" description:"Additional HTTP headers to send with the request"`
	Body         string                 `mapstructure:"body" description:"Request body, overrides \"body-template\""`
	BodyTemplate string                 `mapstructure:"body-template" description:"Path to a file containing the request body (go template). See https://creativeprojects.github.io/resticprofile/configuration/http_hooks/#body-template"`
	SkipTLS      bool                   `mapstructure:"skip-tls-verification" description:"Enables insecure TLS (without verification), see also \"global.ca-certificates\""`
}

// SendMonitoringHeader is used to send HTTP headers
type SendMonitoringHeader struct {
	Name  string            `mapstructure:"name" regex:"^\\w([\\w-]+)\\w$" examples:"\"Authorization\";\"Cache-Control\";\"Content-Disposition\";\"Content-Type\"" description:"Name of the HTTP header"`
	Value ConfidentialValue `mapstructure:"value" examples:"\"Bearer ...\";\"Basic ...\";\"no-cache\";\"attachment;; filename=stats.txt\";\"application/json\";\"text/plain\";\"text/xml\"" description:"Value of the header"`
}

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
		Name:          name,
		config:        c,
		OtherSections: make(map[string]*GenericSection),
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
			if p.config == nil {
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
			} else if p.config != nil {
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

// SetLegacyArg is used to activate the legacy (broken) mode of sending arguments on the restic command line
func (p *Profile) SetLegacyArg(legacy bool) {
	p.legacyArg = legacy
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

func (p *Profile) resolveSourcePath(sourceBase string, sourcePaths ...string) []string {
	var applySourceBase, applyBaseDir pathFix

	// Backup source is NOT relative to the configuration, but to PWD or sourceBase (if not empty)
	// Applying "sourceBase" if set
	if sourceBase = strings.TrimSpace(sourceBase); sourceBase != "" {
		sourceBase = fixPath(sourceBase, expandEnv, expandUserHome)
		applySourceBase = absolutePrefix(sourceBase)
	}
	// Applying a custom PWD eagerly so that own commands (e.g. "show") display correct paths
	if p.BaseDir != "" {
		applyBaseDir = absolutePrefix(p.BaseDir)
	}

	sourcePaths = fixPaths(sourcePaths, expandEnv, expandUserHome, applySourceBase, applyBaseDir)
	sourcePaths = resolveGlob(sourcePaths)
	sourcePaths = fixPaths(sourcePaths, filepath.ToSlash, filepath.FromSlash)
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
	resolvePath := func(origin string, paths []string, revolver func(string) []string) (resolved []string) {
		hasAbsoluteBase := len(p.BaseDir) > 0 && filepath.IsAbs(p.BaseDir)
		for _, path := range paths {
			if len(path) > 0 {
				for _, rp := range revolver(path) {
					if rp != path && p.config != nil && !hasAbsoluteBase {
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
					return fixPaths(p.resolveSourcePath(basePath, path), absolutePath)
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

// GetCommonFlags returns the flags common to all commands
func (p *Profile) GetCommonFlags() (flags *shell.Args) {
	// Flags from the profile fields
	flags = shell.NewArgs().SetLegacyArg(p.legacyArg)
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
	return p.Retention != nil && len(p.Retention.Schedule) > 0
}

// GetBackupSource returns the directories to backup
func (p *Profile) GetBackupSource() []string {
	if p.Backup == nil {
		return nil
	}
	return p.Backup.Source
}

// DefinedCommands returns all commands (also called sections) defined in the profile (backup, check, forget, etc.)
func (p *Profile) DefinedCommands() (commands []string) {
	if commands = maps.Keys(GetSectionsWith[any](p)); commands != nil {
		sort.Strings(commands)
	}
	return
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
func (p *Profile) SchedulableCommands() (commands []string) {
	if commands = maps.Keys(GetDeclaredSectionsWith[Scheduling](p)); commands != nil {
		sort.Strings(commands)
	}
	return
}

// Schedules returns a slice of ScheduleConfig that satisfy the schedule.Config interface
func (p *Profile) Schedules() []*ScheduleConfig {
	// All SectionWithSchedule (backup, check, prune, etc)
	sections := GetSectionsWith[Scheduling](p)
	configs := make([]*ScheduleConfig, 0, len(sections))

	for name, section := range sections {
		if s := section.GetSchedule(); len(s.Schedule) > 0 {
			env := util.NewDefaultEnvironment()

			if len(s.ScheduleEnvCapture) > 0 {
				// Capture OS env
				env.SetValues(os.Environ()...)

				// Capture profile env
				for key, value := range p.Environment {
					env.Put(key, value.Value())
				}

				for index, key := range env.Names() {
					matched := slices.ContainsFunc(s.ScheduleEnvCapture, func(pattern string) bool {
						matched, err := filepath.Match(pattern, key)
						if err != nil && index == 0 {
							clog.Tracef("env not matched with invalid glob expression '%s': %s", pattern, err.Error())
						}
						return matched
					})
					if !matched {
						env.Remove(key)
					}
				}
			}

			config := &ScheduleConfig{
				Title:       p.Name,
				SubTitle:    name,
				Schedules:   s.Schedule,
				Permission:  s.SchedulePermission,
				Environment: env.Values(),
				Log:         s.ScheduleLog,
				LockMode:    s.ScheduleLockMode,
				LockWait:    s.ScheduleLockWait,
				Priority:    s.SchedulePriority,
				ConfigFile:  p.config.configFile,
			}

			if len(config.Log) > 0 {
				if tempDir, err := util.TempDir(); err == nil && strings.HasPrefix(config.Log, filepath.ToSlash(tempDir)) {
					config.Log = path.Join(constants.TemporaryDirMarker, config.Log[len(tempDir):])
				}
			}

			configs = append(configs, config)
		}
	}

	return configs
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
