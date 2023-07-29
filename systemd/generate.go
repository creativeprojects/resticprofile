//go:build !darwin && !windows

package systemd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/spf13/afero"
	"golang.org/x/exp/slices"
)

const (
	defaultPermission = 0644
	systemdSystemDir  = "/etc/systemd/system/"

	systemdUnitDefaultTmpl = `[Unit]
Description={{ .JobDescription }}

[Service]
Type=notify
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .CommandLine }}
{{ if .Nice }}Nice={{ .Nice }}{{ end }}
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

var (
	fs afero.Fs
)

// templateInfo to create systemd unit
type templateInfo struct {
	templates.DefaultData
	JobDescription   string
	TimerDescription string
	WorkingDirectory string
	CommandLine      string
	OnCalendar       []string
	SystemdProfile   string
	Nice             int
	Environment      []string
}

// Config for generating systemd unit and timer files
type Config struct {
	CommandLine      string
	Environment      []string
	WorkingDirectory string
	Title            string
	SubTitle         string
	JobDescription   string
	TimerDescription string
	Schedules        []string
	UnitType         UnitType
	Priority         string
	UnitFile         string
	TimerFile        string
}

func init() {
	fs = afero.NewOsFs()
}

// Generate systemd unit
func Generate(config Config) error {
	var err error
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
	// add $HOME to the environment variables (as a fallback if not defined in profile)
	if home, err := os.UserHomeDir(); err == nil {
		environment = append(environment, fmt.Sprintf("HOME=%s", home))
	}
	// also add $SUDO_USER to env variables
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		environment = append(environment, fmt.Sprintf("SUDO_USER=%s", sudoUser))
	}

	nice := constants.DefaultBackgroundNiceFlag
	if config.Priority == constants.SchedulePriorityStandard {
		nice = constants.DefaultStandardNiceFlag
	}

	info := templateInfo{
		DefaultData:      templates.NewDefaultData(nil),
		JobDescription:   config.JobDescription,
		TimerDescription: config.TimerDescription,
		WorkingDirectory: config.WorkingDirectory,
		CommandLine:      config.CommandLine,
		OnCalendar:       config.Schedules,
		SystemdProfile:   systemdProfile,
		Nice:             nice,
		Environment:      environment,
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
	if err := unitTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName := filepath.Join(systemdUserDir, systemdProfile)
	clog.Infof("writing %v", filePathName)
	if err := afero.WriteFile(fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}
	data.Reset()

	systemdTimerTmpl, err := loadTemplate(config.TimerFile, systemdTimerDefaultTmpl)
	if err != nil {
		return err
	}
	timerTmpl, err := templates.New("timer.unit").Parse(systemdTimerTmpl)
	if err != nil {
		return err
	}
	if err := timerTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName = filepath.Join(systemdUserDir, timerProfile)
	clog.Infof("writing %v", filePathName)
	if err := afero.WriteFile(fs, filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}
	return nil
}

// GetServiceFile returns the service file name for the profile
func GetServiceFile(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.service", commandName, profileName)
}

// GetTimerFile returns the timer file name for the profile
func GetTimerFile(profileName, commandName string) string {
	return fmt.Sprintf("resticprofile-%s@profile-%s.timer", commandName, profileName)
}

// GetUserDir returns the default directory where systemd stores user units
func GetUserDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	systemdUserDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	if err := fs.MkdirAll(systemdUserDir, 0700); err != nil {
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
