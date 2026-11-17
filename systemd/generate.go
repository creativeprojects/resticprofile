//go:build !darwin && !windows && !openbsd && !netbsd && !freebsd

package systemd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/user"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/spf13/afero"
)

const (
	defaultPermission         = 0o644
	environmentFilePermission = 0o600 // the captured environment can contain secrets
	environmentFileName       = "resticprofile.env"
	systemdSystemDir          = "/etc/systemd/system/"

	systemdUnitDefaultTmpl = `[Unit]
Description={{ .JobDescription }}
{{ if .AfterNetworkOnline }}After=network-online.target
{{ end }}
[Service]
Type=notify
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .CommandLine }}
{{ if .Nice }}Nice={{ .Nice }}
{{ end -}}
{{ if .CPUSchedulingPolicy }}CPUSchedulingPolicy={{ .CPUSchedulingPolicy }}
{{ end -}}
{{ if .IOSchedulingClass }}IOSchedulingClass={{ .IOSchedulingClass }}
{{ end -}}
{{ if .IOSchedulingPriority }}IOSchedulingPriority={{ .IOSchedulingPriority }}
{{ end -}}
 {{ if .User }}User={{ .User }}
{{ end -}}
{{ range .Environment -}}
Environment="{{ . }}"
{{ end -}}
{{ if .EnvironmentFile }}EnvironmentFile={{ .EnvironmentFile }}
{{ end -}}
`

	systemdTimerDefaultTmpl = `[Unit]
Description={{ .TimerDescription }}

[Timer]
{{ range .OnCalendar -}}
OnCalendar={{ . }}
{{ end -}}
Unit={{ .SystemdProfile }}
Persistent=true

[Install]
WantedBy=timers.target
`
)

// UnitType is either user or system
type UnitType int

// Type of systemd unit
const (
	UserUnit UnitType = iota
	SystemUnit
)

// templateInfo to create systemd unit
type templateInfo struct {
	templates.DefaultData

	JobDescription       string
	TimerDescription     string
	WorkingDirectory     string
	CommandLine          string
	OnCalendar           []string
	SystemdProfile       string
	Nice                 int
	Environment          []string
	EnvironmentFile      string
	AfterNetworkOnline   bool
	CPUSchedulingPolicy  string
	IOSchedulingClass    int
	IOSchedulingPriority int
	User                 string
}

// Config for generating systemd unit and timer files
type Config struct {
	CommandLine          string
	Environment          []string
	WorkingDirectory     string
	Title                string
	SubTitle             string
	JobDescription       string
	TimerDescription     string
	Schedules            []string
	UnitType             UnitType
	Priority             string // standard or background
	UnitFile             string
	TimerFile            string
	DropInFiles          []string
	AfterNetworkOnline   bool
	Nice                 int
	CPUSchedulingPolicy  string
	IOSchedulingClass    int
	IOSchedulingPriority int
	User                 string
}

type Unit struct {
	fs   afero.Fs
	user user.User
}

func NewUnit(user user.User) Unit {
	return Unit{
		fs:   afero.NewOsFs(),
		user: user,
	}
}

// Generate systemd unit
func (u Unit) Generate(config Config) error {
	var err error
	systemdProfile := GetServiceFile(config.Title, config.SubTitle)
	timerProfile := GetTimerFile(config.Title, config.SubTitle)

	systemdUserDir := systemdSystemDir
	if config.UnitType == UserUnit {
		systemdUserDir, err = u.GetUserDir()
		if err != nil {
			return err
		}
	}

	var environment []string
	if config.UnitType == SystemUnit && config.User == "" {
		// permission = system
		environment = append(environment, fmt.Sprintf("HOME=%s", u.user.SudoHomeDir))
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
			environment = append(environment, fmt.Sprintf("SUDO_USER=%s", sudoUser))
		}
	} else if u.user.UserHomeDir != "" {
		// permission = user or user_logged_on
		environment = append(environment, fmt.Sprintf("HOME=%s", u.user.UserHomeDir))
	}

	// Any user can read Environment= ("systemctl show"): the captured environment, which can contain secrets,
	// goes to a file only readable by the owner of the unit. Generated variables keep precedence.
	capturedEnvironment := collect.All(config.Environment, func(env string) bool {
		return !slices.ContainsFunc(environment, func(generated string) bool { return envName(generated) == envName(env) })
	})
	environmentFilePath := filepath.Join(systemdUserDir, GetServiceFileDropInDir(config.Title, config.SubTitle), environmentFileName)
	environmentFile := ""
	if len(capturedEnvironment) > 0 {
		environmentFile = environmentFilePath
	}

	policy := ""
	if config.Priority == constants.SchedulePriorityBackground {
		policy = "idle"
	}

	info := templateInfo{
		DefaultData:          templates.NewDefaultData(nil),
		JobDescription:       config.JobDescription,
		TimerDescription:     config.TimerDescription,
		WorkingDirectory:     config.WorkingDirectory,
		CommandLine:          config.CommandLine,
		OnCalendar:           config.Schedules,
		AfterNetworkOnline:   config.AfterNetworkOnline,
		SystemdProfile:       systemdProfile,
		Nice:                 config.Nice,
		Environment:          environment,
		EnvironmentFile:      environmentFile,
		CPUSchedulingPolicy:  policy,
		IOSchedulingClass:    config.IOSchedulingClass,
		IOSchedulingPriority: config.IOSchedulingPriority,
		User:                 config.User,
	}

	var data bytes.Buffer

	systemdUnitTmpl, err := u.loadTemplate(config.UnitFile, systemdUnitDefaultTmpl)
	if err != nil {
		return err
	}
	if environmentFile != "" && !strings.Contains(systemdUnitTmpl, ".EnvironmentFile") {
		clog.Warningf("unit template %q doesn't use {{ .EnvironmentFile }}: the service won't get the captured environment", config.UnitFile)
	}
	unitTmpl, err := templates.New("systemd.unit").Parse(systemdUnitTmpl)
	if err != nil {
		return err
	}
	if err = unitTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName := filepath.Join(systemdUserDir, systemdProfile)
	clog.Debugf("writing %v", filePathName)
	if err = afero.WriteFile(u.fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}
	data.Reset()

	if config.UnitType == UserUnit && u.user.Sudo {
		// we need to change the owner to the original account
		_ = u.fs.Chown(filePathName, u.user.Uid, u.user.Gid)
	}

	systemdTimerTmpl, err := u.loadTemplate(config.TimerFile, systemdTimerDefaultTmpl)
	if err != nil {
		return err
	}
	timerTmpl, err := templates.New("timer.unit").Parse(systemdTimerTmpl)
	if err != nil {
		return err
	}
	if err = timerTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName = filepath.Join(systemdUserDir, timerProfile)
	clog.Debugf("writing %v", filePathName)
	if err = afero.WriteFile(u.fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}

	if config.UnitType == UserUnit && u.user.Sudo {
		// we need to change the owner to the original account
		_ = u.fs.Chown(filePathName, u.user.Uid, u.user.Gid)
	}

	// written before the drop-ins, which change the owner of the folder content for user units
	if err = u.writeEnvironmentFile(environmentFilePath, capturedEnvironment); err != nil {
		return err
	}

	existingFiles := collect.All(config.DropInFiles, u.DropInFileExists)

	dropIns := map[string][]string{
		GetTimerFileDropInDir(config.Title, config.SubTitle):   collect.All(existingFiles, u.IsTimerDropIn),
		GetServiceFileDropInDir(config.Title, config.SubTitle): collect.All(existingFiles, collect.Not(u.IsTimerDropIn)),
	}
	for dropInDir, dropInFiles := range dropIns {
		dropInDir = filepath.Join(systemdUserDir, dropInDir)
		if err = u.createDropIns(dropInDir, dropInFiles); err != nil {
			return err
		}
		if config.UnitType == UserUnit && u.user.Sudo {
			// we need to change the owner to the original account
			_ = afero.Walk(u.fs, dropInDir, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}
				_ = u.fs.Chown(path, u.user.Uid, u.user.Gid)
				return nil
			})
		}
	}

	return nil
}

// GetUserDir returns the default directory where systemd stores user units
func (u Unit) GetUserDir() (string, error) {
	systemdUserDir := filepath.Join(u.user.UserHomeDir, ".config", "systemd", "user")
	if err := u.fs.MkdirAll(systemdUserDir, 0o700); err != nil {
		return "", err
	}
	return systemdUserDir, nil
}

// GetSystemDir returns the path where the local systemd units are stored
func GetSystemDir() string {
	return systemdSystemDir
}

// GetServiceFile returns the service file name for the profile
func GetServiceFile(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.service", commandName, profileName)
}

// GetServiceFileDropInDir returns the service file drop-in dir name for the profile
func GetServiceFileDropInDir(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.service.d", commandName, profileName)
}

// GetTimerFileDropInDir returns the timer file drop-in dir name for the profile
func GetTimerFileDropInDir(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.timer.d", commandName, profileName)
}

// GetTimerFile returns the timer file name for the profile
func GetTimerFile(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.timer", commandName, profileName)
}

// loadTemplate loads the content of the filename if the parameter is not empty,
// or returns the default template if the filename parameter is empty
func (u Unit) loadTemplate(filename, defaultTmpl string) (string, error) {
	if filename == "" {
		return defaultTmpl, nil
	}
	clog.Debugf("using template file %q", filename)
	file, err := u.fs.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	output := &strings.Builder{}
	_, err = io.Copy(output, file)
	if err != nil {
		return "", err
	}
	return output.String(), nil
}

// writeEnvironmentFile saves the environment in a file only readable by its owner, or removes the file when there's no environment
func (u Unit) writeEnvironmentFile(filename string, environment []string) error {
	// a new file gets the right permission, and O_EXCL doesn't follow a symlink
	if err := u.fs.Remove(filename); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if len(environment) == 0 {
		return nil
	}
	if err := u.fs.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	content := strings.Builder{}
	for _, env := range environment {
		name, value, _ := strings.Cut(env, "=")
		content.WriteString(name + "=" + quoteEnvironmentValue(value) + "\n")
	}
	clog.Debugf("writing %v", filename)
	file, err := u.fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, environmentFilePermission)
	if err != nil {
		return err
	}
	_, err = file.WriteString(content.String())
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	return err
}

// environmentValueEscaper escapes the only sequences systemd interprets in a double-quoted environment file value
var environmentValueEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`)

func quoteEnvironmentValue(value string) string {
	return `"` + environmentValueEscaper.Replace(value) + `"`
}

func envName(env string) string {
	name, _, _ := strings.Cut(env, "=")
	return name
}
