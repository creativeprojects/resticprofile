package config

import (
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

// Profile contains the whole profile configuration
type Profile struct {
	config        *Config
	Name          string
	Quiet         bool                      `mapstructure:"quiet" argument:"quiet"`
	Verbose       bool                      `mapstructure:"verbose" argument:"verbose"`
	Repository    string                    `mapstructure:"repository" argument:"repo"`
	PasswordFile  string                    `mapstructure:"password-file" argument:"password-file"`
	CacheDir      string                    `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert        string                    `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert string                    `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize    bool                      `mapstructure:"initialize"`
	Inherit       string                    `mapstructure:"inherit"`
	Lock          string                    `mapstructure:"lock"`
	ForceLock     bool                      `mapstructure:"force-inactive-lock"`
	RunBefore     []string                  `mapstructure:"run-before"`
	RunAfter      []string                  `mapstructure:"run-after"`
	RunAfterFail  []string                  `mapstructure:"run-after-fail"`
	StatusFile    string                    `mapstructure:"status-file"`
	Environment   map[string]string         `mapstructure:"env"`
	Backup        *BackupSection            `mapstructure:"backup"`
	Retention     *RetentionSection         `mapstructure:"retention"`
	Check         *OtherSectionWithSchedule `mapstructure:"check"`
	Snapshots     map[string]interface{}    `mapstructure:"snapshots"`
	Forget        map[string]interface{}    `mapstructure:"forget"`
	Mount         map[string]interface{}    `mapstructure:"mount"`
	OtherFlags    map[string]interface{}    `mapstructure:",remain"`
}

// BackupSection contains the specific configuration to the 'backup' command
type BackupSection struct {
	CheckBefore        bool                   `mapstructure:"check-before"`
	CheckAfter         bool                   `mapstructure:"check-after"`
	RunBefore          []string               `mapstructure:"run-before"`
	RunAfter           []string               `mapstructure:"run-after"`
	UseStdin           bool                   `mapstructure:"stdin" argument:"stdin"`
	Source             []string               `mapstructure:"source"`
	ExcludeFile        []string               `mapstructure:"exclude-file" argument:"exclude-file"`
	FilesFrom          []string               `mapstructure:"files-from" argument:"files-from"`
	Schedule           []string               `mapstructure:"schedule"`
	SchedulePermission string                 `mapstructure:"schedule-permission"`
	ScheduleLog        string                 `mapstructure:"schedule-log"`
	OtherFlags         map[string]interface{} `mapstructure:",remain"`
}

// RetentionSection contains the specific configuration to
// the 'forget' command when running as part of a backup
type RetentionSection struct {
	BeforeBackup       bool                   `mapstructure:"before-backup"`
	AfterBackup        bool                   `mapstructure:"after-backup"`
	Schedule           []string               `mapstructure:"schedule"`
	SchedulePermission string                 `mapstructure:"schedule-permission"`
	ScheduleLog        string                 `mapstructure:"schedule-log"`
	OtherFlags         map[string]interface{} `mapstructure:",remain"`
}

// OtherSectionWithSchedule is a section containing schedule only specific parameters
// (the other parameters being for restic)
type OtherSectionWithSchedule struct {
	Schedule           []string               `mapstructure:"schedule"`
	SchedulePermission string                 `mapstructure:"schedule-permission"`
	ScheduleLog        string                 `mapstructure:"schedule-log"`
	OtherFlags         map[string]interface{} `mapstructure:",remain"`
}

// NewProfile instantiates a new blank profile
func NewProfile(c *Config, name string) *Profile {
	return &Profile{
		Name:   name,
		config: c,
	}
}

// SetRootPath changes the path of all the relative paths and files in the configuration
func (p *Profile) SetRootPath(rootPath string) {

	p.Lock = fixPath(p.Lock, expandEnv, absolutePrefix(rootPath), unixSpaces)
	p.PasswordFile = fixPath(p.PasswordFile, expandEnv, absolutePrefix(rootPath), unixSpaces)
	p.CacheDir = fixPath(p.CacheDir, expandEnv, absolutePrefix(rootPath), unixSpaces)
	p.CACert = fixPath(p.CACert, expandEnv, absolutePrefix(rootPath), unixSpaces)
	p.TLSClientCert = fixPath(p.TLSClientCert, expandEnv, absolutePrefix(rootPath), unixSpaces)

	if p.Backup != nil {
		if p.Backup.ExcludeFile != nil && len(p.Backup.ExcludeFile) > 0 {
			p.Backup.ExcludeFile = fixPaths(p.Backup.ExcludeFile, expandEnv, absolutePrefix(rootPath), unixSpaces)
		}

		if p.Backup.FilesFrom != nil && len(p.Backup.FilesFrom) > 0 {
			p.Backup.FilesFrom = fixPaths(p.Backup.FilesFrom, expandEnv, absolutePrefix(rootPath), unixSpaces)
		}

		// Backup source is NOT relative to the configuration, but where the script was launched instead
		if p.Backup.Source != nil && len(p.Backup.Source) > 0 {
			p.Backup.Source = fixPaths(p.Backup.Source, expandEnv, unixSpaces)
		}
	}
}

// SetHost will replace any host value from a boolean to the hostname
func (p *Profile) SetHost(hostname string) {
	if p.Backup != nil && p.Backup.OtherFlags != nil {
		replaceTrueValue(p.Backup.OtherFlags, constants.ParameterHost, hostname)
	}
	if p.Retention != nil && p.Retention.OtherFlags != nil {
		replaceTrueValue(p.Retention.OtherFlags, constants.ParameterHost, hostname)
	}
	if p.Snapshots != nil {
		replaceTrueValue(p.Snapshots, constants.ParameterHost, hostname)
	}
	if p.Forget != nil {
		replaceTrueValue(p.Forget, constants.ParameterHost, hostname)
	}
	if p.Mount != nil {
		replaceTrueValue(p.Mount, constants.ParameterHost, hostname)
	}
}

// GetCommonFlags returns the flags common to all commands
func (p *Profile) GetCommonFlags() map[string][]string {
	// Flags from the profile fields
	flags := convertStructToFlags(*p)

	flags = addOtherFlags(flags, p.OtherFlags)

	return flags
}

// GetCommandFlags returns the flags specific to the command (backup, snapshots, forget, etc.)
func (p *Profile) GetCommandFlags(command string) map[string][]string {
	flags := p.GetCommonFlags()

	switch command {
	case constants.CommandBackup:
		if p.Backup == nil {
			clog.Warning("No definition for backup command in this profile")
			break
		}
		commandFlags := convertStructToFlags(*p.Backup)
		if len(commandFlags) > 0 {
			flags = mergeFlags(flags, commandFlags)
		}
		flags = addOtherFlags(flags, p.Backup.OtherFlags)

	case constants.CommandSnapshots:
		if p.Snapshots != nil {
			flags = addOtherFlags(flags, p.Snapshots)
		}

	case constants.CommandCheck:
		if p.Check != nil && p.Check.OtherFlags != nil {
			flags = addOtherFlags(flags, p.Check.OtherFlags)
		}

	case constants.CommandForget:
		if p.Forget != nil {
			flags = addOtherFlags(flags, p.Forget)
		}

	case constants.CommandMount:
		if p.Mount != nil {
			flags = addOtherFlags(flags, p.Mount)
		}
	}

	return flags
}

// GetRetentionFlags returns the flags specific to the "forget" command being run as part of a backup
func (p *Profile) GetRetentionFlags() map[string][]string {
	// if there was no "other" flags, the map could be un-initialized
	if p.Retention.OtherFlags == nil {
		p.Retention.OtherFlags = make(map[string]interface{})
	}

	flags := p.GetCommonFlags()
	// Special case of retention: we do copy the "source" from "backup" as "path" if it hasn't been redefined in "retention"
	if _, found := p.Retention.OtherFlags[constants.ParameterPath]; !found {
		p.Retention.OtherFlags[constants.ParameterPath] = fixPaths(p.Backup.Source, absolutePath)
	}
	flags = addOtherFlags(flags, p.Retention.OtherFlags)
	return flags
}

// GetBackupSource returns the directories to backup
func (p *Profile) GetBackupSource() []string {
	if p.Backup == nil {
		return nil
	}
	return p.Backup.Source
}

// Schedules returns a slice of ScheduleConfig that satisfy the schedule.Config interface
func (p *Profile) Schedules() []*ScheduleConfig {
	// Default to 3: backup, retention and check
	configs := make([]*ScheduleConfig, 0, 3)
	// Backup
	if p.Backup != nil && p.Backup.Schedule != nil && len(p.Backup.Schedule) > 0 {
		config := &ScheduleConfig{
			profileName: p.Name,
			commandName: constants.CommandBackup,
			schedules:   p.Backup.Schedule,
			permission:  p.Backup.SchedulePermission,
			environment: p.Environment,
			nice:        10, // hard-coded for now
			logfile:     p.Backup.ScheduleLog,
		}
		configs = append(configs, config)
	}
	// Retention (forget)
	if p.Retention != nil && p.Retention.Schedule != nil && len(p.Retention.Schedule) > 0 {
		config := &ScheduleConfig{
			profileName: p.Name,
			commandName: constants.SectionConfigurationRetention,
			schedules:   p.Retention.Schedule,
			permission:  p.Retention.SchedulePermission,
			environment: p.Environment,
			nice:        10, // hard-coded for now
			logfile:     p.Retention.ScheduleLog,
		}
		configs = append(configs, config)
	}
	// Check
	if p.Check != nil && p.Check.Schedule != nil && len(p.Check.Schedule) > 0 {
		config := &ScheduleConfig{
			profileName: p.Name,
			commandName: constants.CommandCheck,
			schedules:   p.Check.Schedule,
			permission:  p.Check.SchedulePermission,
			environment: p.Environment,
			nice:        10, // hard-coded for now
			logfile:     p.Check.ScheduleLog,
		}
		configs = append(configs, config)
	}
	return configs
}
func addOtherFlags(flags map[string][]string, otherFlags map[string]interface{}) map[string][]string {
	if len(otherFlags) == 0 {
		return flags
	}

	// Add other flags
	for name, value := range otherFlags {
		if convert, ok := stringifyValueOf(value); ok {
			flags[name] = convert
		}
	}
	return flags
}

func mergeFlags(flags, newFlags map[string][]string) map[string][]string {
	if len(flags) == 0 && newFlags != nil {
		return newFlags
	}
	if flags != nil && len(newFlags) == 0 {
		return flags
	}
	for key, value := range newFlags {
		flags[key] = value
	}
	return flags
}

func replaceTrueValue(source map[string]interface{}, key, replace string) {
	if genericValue, ok := source[key]; ok {
		if value, ok := genericValue.(bool); ok {
			if value {
				source[key] = replace
			}
		}
	}
}
