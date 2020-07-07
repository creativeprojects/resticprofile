package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/creativeprojects/resticprofile/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

// Profile contains the whole profile configuration
type Profile struct {
	config        *Config
	Name          string
	Quiet         bool                   `mapstructure:"quiet" argument:"quiet"`
	Verbose       bool                   `mapstructure:"verbose" argument:"verbose"`
	Repository    string                 `mapstructure:"repository" argument:"repo"`
	PasswordFile  string                 `mapstructure:"password-file" argument:"password-file"`
	CacheDir      string                 `mapstructure:"cache-dir" argument:"cache-dir"`
	CACert        string                 `mapstructure:"cacert" argument:"cacert"`
	TLSClientCert string                 `mapstructure:"tls-client-cert" argument:"tls-client-cert"`
	Initialize    bool                   `mapstructure:"initialize"`
	Inherit       string                 `mapstructure:"inherit"`
	Lock          string                 `mapstructure:"lock"`
	RunBefore     []string               `mapstructure:"run-before"`
	RunAfter      []string               `mapstructure:"run-after"`
	RunAfterFail  []string               `mapstructure:"run-after-fail"`
	Schedule      []string               `mapstructure:"schedule"`
	ScheduleType  string                 `mapstructure:"schedule-type"`
	Environment   map[string]string      `mapstructure:"env"`
	Backup        *BackupSection         `mapstructure:"backup"`
	Retention     *RetentionSection      `mapstructure:"retention"`
	Snapshots     map[string]interface{} `mapstructure:"snapshots"`
	Forget        map[string]interface{} `mapstructure:"forget"`
	Check         map[string]interface{} `mapstructure:"check"`
	Mount         map[string]interface{} `mapstructure:"mount"`
	OtherFlags    map[string]interface{} `mapstructure:",remain"`
}

// BackupSection contains the specific configuration to the 'backup' command
type BackupSection struct {
	CheckBefore bool                   `mapstructure:"check-before"`
	CheckAfter  bool                   `mapstructure:"check-after"`
	RunBefore   []string               `mapstructure:"run-before"`
	RunAfter    []string               `mapstructure:"run-after"`
	UseStdin    bool                   `mapstructure:"stdin" argument:"stdin"`
	Source      []string               `mapstructure:"source"`
	ExcludeFile []string               `mapstructure:"exclude-file" argument:"exclude-file"`
	FilesFrom   []string               `mapstructure:"files-from" argument:"files-from"`
	OtherFlags  map[string]interface{} `mapstructure:",remain"`
}

// RetentionSection contains the specific configuration to the 'forget' command run as part of a backup
type RetentionSection struct {
	BeforeBackup bool                   `mapstructure:"before-backup"`
	AfterBackup  bool                   `mapstructure:"after-backup"`
	OtherFlags   map[string]interface{} `mapstructure:",remain"`
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

	p.Lock = fixPath(p.Lock, rootPath)
	p.PasswordFile = fixPath(p.PasswordFile, rootPath)
	p.CacheDir = fixPath(p.CacheDir, rootPath)
	p.CACert = fixPath(p.CACert, rootPath)
	p.TLSClientCert = fixPath(p.TLSClientCert, rootPath)

	if p.Backup != nil {
		if p.Backup.ExcludeFile != nil && len(p.Backup.ExcludeFile) > 0 {
			for i := 0; i < len(p.Backup.ExcludeFile); i++ {
				p.Backup.ExcludeFile[i] = fixPath(p.Backup.ExcludeFile[i], rootPath)
			}
		}

		if p.Backup.FilesFrom != nil && len(p.Backup.FilesFrom) > 0 {
			for i := 0; i < len(p.Backup.FilesFrom); i++ {
				p.Backup.FilesFrom[i] = fixPath(p.Backup.FilesFrom[i], rootPath)
			}
		}

		// Do we need to do source files? (it wasn't the case before v0.6.0)
		if p.Backup.Source != nil && len(p.Backup.Source) > 0 {
			for i := 0; i < len(p.Backup.Source); i++ {
				p.Backup.Source[i] = fixPath(p.Backup.Source[i], rootPath)
			}
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

// GetCommandFlags returns the flags specific to the command (backup, snapshots, etc.)
func (p *Profile) GetCommandFlags(command string) map[string][]string {
	flags := p.GetCommonFlags()

	switch command {
	case constants.CommandBackup:
		if p.Backup == nil {
			clog.Warning("No definition for backup command in this profile")
			break
		}
		commandFlags := convertStructToFlags(*p.Backup)
		if commandFlags != nil && len(commandFlags) > 0 {
			flags = mergeFlags(flags, commandFlags)
		}
		flags = addOtherFlags(flags, p.Backup.OtherFlags)

	case constants.CommandSnapshots:
		if p.Snapshots != nil {
			flags = addOtherFlags(flags, p.Snapshots)
		}

	case constants.CommandCheck:
		if p.Check != nil {
			flags = addOtherFlags(flags, p.Check)
		}
	}

	return flags
}

// GetRetentionFlags returns the flags specific to the "forget" command being run as part of a backup
func (p *Profile) GetRetentionFlags() map[string][]string {
	flags := p.GetCommonFlags()
	// Special case of retention: we do copy the "source" from "backup" as "path" if it hasn't been redefined in "retention"
	if _, found := p.Retention.OtherFlags[constants.ParameterPath]; !found {
		p.Retention.OtherFlags[constants.ParameterPath] = p.Backup.Source
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

func addOtherFlags(flags map[string][]string, otherFlags map[string]interface{}) map[string][]string {
	if otherFlags == nil || len(otherFlags) == 0 {
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
	if (flags == nil || len(flags) == 0) && newFlags != nil {
		return newFlags
	}
	if flags != nil && (newFlags == nil || len(newFlags) == 0) {
		return flags
	}
	for key, value := range newFlags {
		flags[key] = value
	}
	return flags
}

func fixPath(source, prefix string) string {
	if strings.Contains(source, "$") || strings.Contains(source, "%") {
		source = os.ExpandEnv(source)
	}
	if runtime.GOOS != "windows" {
		prefix = strings.ReplaceAll(strings.TrimSpace(prefix), " ", `\ `)
		source = strings.ReplaceAll(strings.TrimSpace(source), " ", `\ `)
	}
	if source == "" ||
		filepath.IsAbs(source) ||
		strings.HasPrefix(source, "~") ||
		strings.HasPrefix(source, "$") ||
		strings.HasPrefix(source, "%") {
		return source
	}
	return filepath.Join(prefix, source)
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
