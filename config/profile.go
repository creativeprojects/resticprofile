package config

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/shell"
)

type Empty interface {
	IsEmpty() bool
}

// Profile contains the whole profile configuration
type Profile struct {
	config               *Config
	legacyArg            bool
	Name                 string
	Description          string                       `mapstructure:"description"`
	Quiet                bool                         `mapstructure:"quiet" argument:"quiet"`
	Verbose              bool                         `mapstructure:"verbose" argument:"verbose"`
	Repository           ConfidentialValue            `mapstructure:"repository" argument:"repo"`
	RepositoryFile       string                       `mapstructure:"repository-file" argument:"repository-file"`
	PasswordFile         string                       `mapstructure:"password-file" argument:"password-file"`
	CacheDir             string                       `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert               string                       `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert        string                       `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize           bool                         `mapstructure:"initialize"`
	Inherit              string                       `mapstructure:"inherit" show:"noshow"`
	Lock                 string                       `mapstructure:"lock"`
	ForceLock            bool                         `mapstructure:"force-inactive-lock"`
	RunBefore            []string                     `mapstructure:"run-before"`
	RunAfter             []string                     `mapstructure:"run-after"`
	RunAfterFail         []string                     `mapstructure:"run-after-fail"`
	RunFinally           []string                     `mapstructure:"run-finally"`
	SendBefore           []SendMonitorSection         `mapstructure:"send-before"`
	SendAfter            []SendMonitorSection         `mapstructure:"send-after"`
	SendAfterFail        []SendMonitorSection         `mapstructure:"send-after-fail"`
	SendFinally          []SendMonitorSection         `mapstructure:"send-finally"`
	StreamError          []StreamErrorSection         `mapstructure:"stream-error"`
	StatusFile           string                       `mapstructure:"status-file"`
	PrometheusSaveToFile string                       `mapstructure:"prometheus-save-to-file"`
	PrometheusPush       string                       `mapstructure:"prometheus-push"`
	PrometheusLabels     map[string]string            `mapstructure:"prometheus-labels"`
	HealthChecksURL      string                       `mapstructure:"healthchecks-url"`
	HealthChecksTimeout  time.Duration                `mapstructure:"healthchecks-timeout"`
	OtherFlags           map[string]interface{}       `mapstructure:",remain"`
	Environment          map[string]ConfidentialValue `mapstructure:"env"`
	Backup               *BackupSection               `mapstructure:"backup"`
	Retention            *RetentionSection            `mapstructure:"retention"`
	Check                *SectionWithSchedule         `mapstructure:"check"`
	Prune                *SectionWithSchedule         `mapstructure:"prune"`
	Snapshots            map[string]interface{}       `mapstructure:"snapshots"`
	Forget               *SectionWithSchedule         `mapstructure:"forget"`
	Mount                map[string]interface{}       `mapstructure:"mount"`
	Copy                 *CopySection                 `mapstructure:"copy"`
	Dump                 map[string]interface{}       `mapstructure:"dump"`
	Find                 map[string]interface{}       `mapstructure:"find"`
	Ls                   map[string]interface{}       `mapstructure:"ls"`
	Restore              map[string]interface{}       `mapstructure:"restore"`
	Stats                map[string]interface{}       `mapstructure:"stats"`
	Tag                  map[string]interface{}       `mapstructure:"tag"`
}

// BackupSection contains the specific configuration to the 'backup' command
type BackupSection struct {
	ScheduleBaseSection `mapstructure:",squash"`
	CheckBefore         bool                   `mapstructure:"check-before"`
	CheckAfter          bool                   `mapstructure:"check-after"`
	RunBefore           []string               `mapstructure:"run-before"`
	RunAfter            []string               `mapstructure:"run-after"`
	RunFinally          []string               `mapstructure:"run-finally"`
	SendBefore          []SendMonitorSection   `mapstructure:"send-before"`
	SendAfter           []SendMonitorSection   `mapstructure:"send-after"`
	SendFinally         []SendMonitorSection   `mapstructure:"send-finally"`
	UseStdin            bool                   `mapstructure:"stdin" argument:"stdin"`
	StdinCommand        []string               `mapstructure:"stdin-command"`
	Source              []string               `mapstructure:"source"`
	Exclude             []string               `mapstructure:"exclude" argument:"exclude" argument-type:"no-glob"`
	Iexclude            []string               `mapstructure:"iexclude" argument:"iexclude" argument-type:"no-glob"`
	ExcludeFile         []string               `mapstructure:"exclude-file" argument:"exclude-file"`
	FilesFrom           []string               `mapstructure:"files-from" argument:"files-from"`
	ExtendedStatus      bool                   `mapstructure:"extended-status" argument:"json"`
	NoErrorOnWarning    bool                   `mapstructure:"no-error-on-warning"`
	HealthChecksUUID    string                 `mapstructure:"healthchecks-uuid"`
	OtherFlags          map[string]interface{} `mapstructure:",remain"`
}

func (s *BackupSection) IsEmpty() bool { return s == nil }

// RetentionSection contains the specific configuration to
// the 'forget' command when running as part of a backup
type RetentionSection struct {
	ScheduleBaseSection `mapstructure:",squash"`
	BeforeBackup        bool                   `mapstructure:"before-backup"`
	AfterBackup         bool                   `mapstructure:"after-backup"`
	OtherFlags          map[string]interface{} `mapstructure:",remain"`
}

func (s *RetentionSection) IsEmpty() bool { return s == nil }

// SectionWithSchedule is a section containing schedule only specific parameters
// (the other parameters being for restic)
type SectionWithSchedule struct {
	ScheduleBaseSection `mapstructure:",squash"`
	OtherFlags          map[string]interface{} `mapstructure:",remain"`
}

func (s *SectionWithSchedule) IsEmpty() bool { return s == nil }

// ScheduleBaseSection contains the parameters for scheduling a command (backup, check, forget, etc.)
type ScheduleBaseSection struct {
	Schedule           []string      `mapstructure:"schedule" show:"noshow"`
	SchedulePermission string        `mapstructure:"schedule-permission" show:"noshow"`
	ScheduleLog        string        `mapstructure:"schedule-log" show:"noshow"`
	SchedulePriority   string        `mapstructure:"schedule-priority" show:"noshow"`
	ScheduleLockMode   string        `mapstructure:"schedule-lock-mode" show:"noshow"`
	ScheduleLockWait   time.Duration `mapstructure:"schedule-lock-wait" show:"noshow"`
}

// CopySection contains the destination parameters for a copy command
type CopySection struct {
	Initialize          bool              `mapstructure:"initialize"`
	Repository          ConfidentialValue `mapstructure:"repository" argument:"repo2"`
	RepositoryFile      string            `mapstructure:"repository-file" argument:"repository-file2"`
	PasswordFile        string            `mapstructure:"password-file" argument:"password-file2"`
	PasswordCommand     string            `mapstructure:"password-command" argument:"password-command2"`
	KeyHint             string            `mapstructure:"key-hint" argument:"key-hint2"`
	ScheduleBaseSection `mapstructure:",squash"`
	OtherFlags          map[string]interface{} `mapstructure:",remain"`
}

func (s *CopySection) IsEmpty() bool { return s == nil }

type StreamErrorSection struct {
	Pattern    string `mapstructure:"pattern"`
	MinMatches int    `mapstructure:"min-matches"`
	MaxRuns    int    `mapstructure:"max-runs"`
	Run        string `mapstructure:"run"`
}

// SendMonitorSection is used to send monitoring information to third party software
type SendMonitorSection struct {
	Method  string              `mapstructure:"method"`
	URL     string              `mapstructure:"url"`
	Headers []SendMonitorHeader `mapstructure:"headers"`
	Body    string              `mapstructure:"body"`
}

// SendMonitorHeader is used to send HTTP headers
type SendMonitorHeader struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
}

// NewProfile instantiates a new blank profile
func NewProfile(c *Config, name string) *Profile {
	return &Profile{
		Name:   name,
		config: c,
	}
}

// ResolveConfiguration resolves dependencies between profile config flags
func (p *Profile) ResolveConfiguration() {
	if p.Backup != nil {
		// Ensure UseStdin is set when Backup.StdinCommand is defined
		if len(p.Backup.StdinCommand) > 0 {
			p.Backup.UseStdin = true
		}

		// Special cases of retention
		if p.Retention != nil {
			if p.Retention.OtherFlags == nil {
				p.Retention.OtherFlags = make(map[string]interface{})
			}
			// Copy "source" from "backup" as "path" if it hasn't been redefined
			if _, found := p.Retention.OtherFlags[constants.ParameterPath]; !found {
				p.Retention.OtherFlags[constants.ParameterPath] = true
			}

			// Copy "tag" from "backup" if it hasn't been redefined (only for Version >= 2 to be backward compatible)
			if p.config != nil && p.config.version >= Version02 {
				if _, found := p.Retention.OtherFlags[constants.ParameterTag]; !found {
					p.Retention.OtherFlags[constants.ParameterTag] = true
				}
			}
		}

		// Copy tags from backup if tag is set to boolean true
		if tags, ok := stringifyValueOf(p.Backup.OtherFlags[constants.ParameterTag]); ok {
			p.SetTag(tags...)
		}

		// Copy parameter path from backup sources if path is set to boolean true
		p.SetPath(p.Backup.Source...)
	} else {
		// Resolve path parameter (no copy since backup is not defined)
		p.SetPath()
	}
}

// SetLegacyArg is used to activate the legacy (broken) mode of sending arguments on the restic command line
func (p *Profile) SetLegacyArg(legacy bool) {
	p.legacyArg = legacy
}

// SetRootPath changes the path of all the relative paths and files in the configuration
func (p *Profile) SetRootPath(rootPath string) {
	p.Lock = fixPath(p.Lock, expandEnv, absolutePrefix(rootPath))
	p.PasswordFile = fixPath(p.PasswordFile, expandEnv, absolutePrefix(rootPath))
	p.RepositoryFile = fixPath(p.RepositoryFile, expandEnv, absolutePrefix(rootPath))
	p.CacheDir = fixPath(p.CacheDir, expandEnv, absolutePrefix(rootPath))
	p.CACert = fixPath(p.CACert, expandEnv, absolutePrefix(rootPath))
	p.TLSClientCert = fixPath(p.TLSClientCert, expandEnv, absolutePrefix(rootPath))

	if p.Backup != nil {
		if p.Backup.ExcludeFile != nil && len(p.Backup.ExcludeFile) > 0 {
			p.Backup.ExcludeFile = fixPaths(p.Backup.ExcludeFile, expandEnv, absolutePrefix(rootPath))
		}

		if p.Backup.FilesFrom != nil && len(p.Backup.FilesFrom) > 0 {
			p.Backup.FilesFrom = fixPaths(p.Backup.FilesFrom, expandEnv, absolutePrefix(rootPath))
		}

		p.Backup.Source = p.resolveSourcePath(p.Backup.Source...)

		if p.Backup.Exclude != nil && len(p.Backup.Exclude) > 0 {
			p.Backup.Exclude = fixPaths(p.Backup.Exclude, expandEnv)
		}

		if p.Backup.Iexclude != nil && len(p.Backup.Iexclude) > 0 {
			p.Backup.Iexclude = fixPaths(p.Backup.Iexclude, expandEnv)
		}
	}

	if p.Copy != nil {
		p.Copy.PasswordFile = fixPath(p.Copy.PasswordFile, expandEnv, absolutePrefix(rootPath))
		p.Copy.RepositoryFile = fixPath(p.Copy.RepositoryFile, expandEnv, absolutePrefix(rootPath))
	}

	// Handle dynamic flags dealing with paths that are relative to root path
	filepathFlags := []string{
		"cacert",
		"tls-client-cert",
		"cache-dir",
		"repository-file",
		"password-file",
	}
	for _, section := range p.allFlagsSections() {
		for _, flag := range filepathFlags {
			if paths, ok := stringifyValueOf(section[flag]); ok && len(paths) > 0 {
				for i, path := range paths {
					if len(path) > 0 {
						paths[i] = fixPath(path, expandEnv, absolutePrefix(rootPath))
					}
				}
				section[flag] = paths
			}
		}
	}
}

func (p *Profile) resolveSourcePath(sourcePaths ...string) []string {
	if len(sourcePaths) > 0 {
		// Backup source is NOT relative to the configuration, but where the script was launched instead
		sourcePaths = fixPaths(sourcePaths, expandEnv, expandUserHome)
		sourcePaths = resolveGlob(sourcePaths)
	}
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
func (p *Profile) SetPath(sourcePaths ...string) {
	resolvePath := func(origin string, paths []string, revolver func(string) []string) (resolved []string) {
		for _, path := range paths {
			if len(path) > 0 {
				for _, rp := range revolver(path) {
					if rp != path && p.config != nil {
						if p.config.issues.changedPaths == nil {
							p.config.issues.changedPaths = make(map[string][]string)
						}
						key := fmt.Sprintf(`%s "%s"`, origin, path)
						p.config.issues.changedPaths[key] = append(p.config.issues.changedPaths[key], rp)
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
					return fixPaths(p.resolveSourcePath(path), absolutePath)
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

func (p *Profile) allFlagsSections() (sections []map[string]interface{}) {
	for _, section := range p.allSections() {
		if flags := p.getSectionOtherFlags(section); flags != nil {
			sections = append(sections, flags)
		}
	}
	return
}

func (p *Profile) getSectionOtherFlags(section interface{}) map[string]interface{} {
	if !reflect.ValueOf(section).IsNil() {
		switch v := section.(type) {
		case *BackupSection:
			return v.OtherFlags
		case *CopySection:
			return v.OtherFlags
		case *SectionWithSchedule:
			return v.OtherFlags
		case *RetentionSection:
			return v.OtherFlags
		case map[string]interface{}:
			return v
		}
	}
	return nil
}

// GetCommonFlags returns the flags common to all commands
func (p *Profile) GetCommonFlags() *shell.Args {
	// Flags from the profile fields
	flags := convertStructToArgs(*p, shell.NewArgs().SetLegacyArg(p.legacyArg))

	flags = addOtherArgs(flags, p.OtherFlags)

	return flags
}

// GetCommandFlags returns the flags specific to the command (backup, snapshots, forget, etc.)
func (p *Profile) GetCommandFlags(command string) *shell.Args {
	flags := p.GetCommonFlags()

	switch command {
	case constants.CommandBackup:
		if p.Backup == nil {
			clog.Warning("No definition for backup command in this profile")
			break
		}
		flags = convertStructToArgs(*p.Backup, flags)

	case constants.CommandCopy:
		if p.Copy != nil {
			flags = convertStructToArgs(*p.Copy, flags)
		}
	}

	// Add generic section flags
	if section := p.allSections()[command]; section != nil {
		flags = addOtherArgs(flags, p.getSectionOtherFlags(section))
	}

	return flags
}

// GetRetentionFlags returns the flags specific to the "forget" command being run as part of a backup
func (p *Profile) GetRetentionFlags() *shell.Args {
	// it shouldn't happen when started as a command, but can occur in a unit test
	if p.Retention == nil {
		return shell.NewArgs()
	}

	flags := p.GetCommonFlags()
	flags = addOtherArgs(flags, p.Retention.OtherFlags)
	return flags
}

// HasDeprecatedRetentionSchedule indicates if there's one or more schedule parameters in the retention section,
// which is deprecated as of 0.11.0
func (p *Profile) HasDeprecatedRetentionSchedule() bool {
	if p.Retention == nil {
		return false
	}
	if len(p.Retention.Schedule) > 0 {
		return true
	}
	return false
}

// GetBackupSource returns the directories to backup
func (p *Profile) GetBackupSource() []string {
	if p.Backup == nil {
		return nil
	}
	return p.Backup.Source
}

// DefinedCommands returns all commands (also called sections) defined in the profile (backup, check, forget, etc.)
func (p *Profile) DefinedCommands() []string {
	sections := p.allSections()
	commands := make([]string, 0, len(sections))
	for name, section := range sections {
		if !reflect.ValueOf(section).IsNil() {
			commands = append(commands, name)
		}
	}
	sort.Strings(commands)
	return commands
}

func (p *Profile) allSections() map[string]interface{} {
	return map[string]interface{}{
		constants.CommandBackup:                 p.Backup,
		constants.CommandCheck:                  p.Check,
		constants.CommandCopy:                   p.Copy,
		constants.CommandDump:                   p.Dump,
		constants.CommandForget:                 p.Forget,
		constants.CommandFind:                   p.Find,
		constants.CommandLs:                     p.Ls,
		constants.CommandMount:                  p.Mount,
		constants.CommandPrune:                  p.Prune,
		constants.CommandRestore:                p.Restore,
		constants.CommandSnapshots:              p.Snapshots,
		constants.CommandStats:                  p.Stats,
		constants.CommandTag:                    p.Tag,
		constants.SectionConfigurationRetention: p.Retention,
	}
}

// SchedulableCommands returns all command names that could have a schedule
func (p *Profile) SchedulableCommands() []string {
	sections := p.allSchedulableSections()
	commands := make([]string, 0, len(sections))
	for name := range sections {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}

// Schedules returns a slice of ScheduleConfig that satisfy the schedule.Config interface
func (p *Profile) Schedules() []*ScheduleConfig {
	// All SectionWithSchedule (backup, check, prune, etc)
	sections := p.allSchedulableSections()
	configs := make([]*ScheduleConfig, 0, len(sections))

	for name, section := range sections {
		if s, ok := getScheduleSection(section); ok && s != nil && s.Schedule != nil && len(s.Schedule) > 0 {
			env := map[string]string{}
			for key, value := range p.Environment {
				env[key] = value.Value()
			}

			config := &ScheduleConfig{
				profileName: p.Name,
				commandName: name,
				schedules:   s.Schedule,
				permission:  s.SchedulePermission,
				environment: env,
				logfile:     s.ScheduleLog,
				lockMode:    s.ScheduleLockMode,
				lockWait:    s.ScheduleLockWait,
				priority:    s.SchedulePriority,
				configfile:  p.config.configFile,
			}

			configs = append(configs, config)
		}
	}

	return configs
}

func (p *Profile) allSchedulableSections() map[string]interface{} {
	sections := p.allSections()
	for name, section := range sections {
		if _, schedulable := getScheduleSection(section); !schedulable {
			delete(sections, name)
		}
	}
	return sections
}

func getScheduleSection(section interface{}) (schedule *ScheduleBaseSection, schedulable bool) {
	switch v := section.(type) {
	case *BackupSection:
		schedulable = true
		if v != nil {
			schedule = &v.ScheduleBaseSection
		}
	case *CopySection:
		schedulable = true
		if v != nil {
			schedule = &v.ScheduleBaseSection
		}
	case *RetentionSection:
		schedulable = true
		if v != nil {
			schedule = &v.ScheduleBaseSection
		}
	case *SectionWithSchedule:
		schedulable = true
		if v != nil {
			schedule = &v.ScheduleBaseSection
		}
	}
	return
}

func replaceTrueValue(source map[string]interface{}, key string, replace ...string) {
	if genericValue, ok := source[key]; ok {
		if value, ok := genericValue.(bool); ok {
			if value {
				source[key] = replace
			}
		}
	}
}
