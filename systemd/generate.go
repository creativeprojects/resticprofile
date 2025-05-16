//go:build !darwin && !windows

package systemd

import (
	"bytes"
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
	defaultPermission = 0o644
	systemdSystemDir  = "/etc/systemd/system/"

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

	environment := slices.Clone(config.Environment)

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

	existingFiles := collect.All(config.DropInFiles, u.FileExists)

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
