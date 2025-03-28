//go:build !darwin && !windows

package systemd

import (
	"bytes"
	"fmt"
	"io"
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

var fs afero.Fs

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

func init() {
	fs = afero.NewOsFs()
}

// Generate systemd unit
func Generate(config Config) error {
	var err error
	u := user.Current()
	systemdProfile := GetServiceFile(config.Title, config.SubTitle)
	timerProfile := GetTimerFile(config.Title, config.SubTitle)

	systemdUserDir := systemdSystemDir
	if config.UnitType == UserUnit {
		systemdUserDir, err = GetUserDir()
		if err != nil {
			return err
		}
	}

	environment := slices.Clone(config.Environment)
	if config.User != "" {
		// resticprofile will start under config.User (user background mode)
		if u.HomeDir != "" {
			environment = append(environment, fmt.Sprintf("HOME=%s", u.HomeDir))
		}
	} else {
		// running resticprofile as root
		if home, err := os.UserHomeDir(); err == nil {
			environment = append(environment, fmt.Sprintf("HOME=%s", home))
		}
		if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
			environment = append(environment, fmt.Sprintf("SUDO_USER=%s", sudoUser))
		}
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

	systemdUnitTmpl, err := loadTemplate(config.UnitFile, systemdUnitDefaultTmpl)
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
	if err = afero.WriteFile(fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}
	data.Reset()

	if config.UnitType == UserUnit && u.SudoRoot {
		// we need to change the owner to the original account
		_ = fs.Chown(filePathName, u.Uid, u.Gid)
	}

	systemdTimerTmpl, err := loadTemplate(config.TimerFile, systemdTimerDefaultTmpl)
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
	if err = afero.WriteFile(fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}

	if config.UnitType == UserUnit && u.SudoRoot {
		// we need to change the owner to the original account
		_ = fs.Chown(filePathName, u.Uid, u.Gid)
	}

	dropIns := map[string][]string{
		GetTimerFileDropInDir(config.Title, config.SubTitle):   collect.All(config.DropInFiles, IsTimerDropIn),
		GetServiceFileDropInDir(config.Title, config.SubTitle): collect.All(config.DropInFiles, collect.Not(IsTimerDropIn)),
	}
	for dropInDir, dropInFiles := range dropIns {
		dropInDir = filepath.Join(systemdUserDir, dropInDir)
		if err = createDropIns(dropInDir, dropInFiles); err != nil {
			return err
		}
		if config.UnitType == UserUnit && u.SudoRoot {
			// we need to change the owner to the original account
			_ = fs.Chown(dropInDir, u.Uid, u.Gid)
		}
	}

	return nil
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

// GetUserDir returns the default directory where systemd stores user units
func GetUserDir() (string, error) {
	u := user.Current()

	systemdUserDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	if err := fs.MkdirAll(systemdUserDir, 0o700); err != nil {
		return "", err
	}
	return systemdUserDir, nil
}

// GetSystemDir returns the path where the local systemd units are stored
func GetSystemDir() string {
	return systemdSystemDir
}

// loadTemplate loads the content of the filename if the parameter is not empty,
// or returns the default template if the filename parameter is empty
func loadTemplate(filename, defaultTmpl string) (string, error) {
	if filename == "" {
		return defaultTmpl, nil
	}
	clog.Debugf("using template file %q", filename)
	file, err := fs.Open(filename)
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
