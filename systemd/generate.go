//+build !darwin,!windows

package systemd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/constants"
)

const (
	defaultPermission = 0644
	systemdSystemDir  = "/etc/systemd/system/"

	systemdUnitBackupUnitTmpl = `[Unit]
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

	systemdUnitBackupTimerTmpl = `[Unit]
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

// TemplateInfo to create systemd unit
type TemplateInfo struct {
	JobDescription   string
	TimerDescription string
	WorkingDirectory string
	CommandLine      string
	OnCalendar       []string
	SystemdProfile   string
	Nice             int
	Environment      []string
}

// Generate systemd unit
func Generate(commandLine, wd, title, subTitle, jobDescription, timerDescription string, onCalendar []string, unitType UnitType, priority string) error {
	var err error
	systemdProfile := GetServiceFile(title, subTitle)
	timerProfile := GetTimerFile(title, subTitle)

	systemdUserDir := systemdSystemDir
	if unitType == UserUnit {
		systemdUserDir, err = GetUserDir()
		if err != nil {
			return err
		}
	}

	environment := make([]string, 0, 2)
	// add $HOME to the environment variables (as a fallback if not defined in profile)
	if home, err := os.UserHomeDir(); err == nil {
		environment = append(environment, fmt.Sprintf("HOME=%s", home))
	}
	// also add $SUDO_USER to env variables
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		environment = append(environment, fmt.Sprintf("SUDO_USER=%s", sudoUser))
	}

	nice := constants.DefaultBackgroundNiceFlag
	if priority == constants.SchedulePriorityStandard {
		nice = constants.DefaultStandardNiceFlag
	}

	info := TemplateInfo{
		JobDescription:   jobDescription,
		TimerDescription: timerDescription,
		WorkingDirectory: wd,
		CommandLine:      commandLine,
		OnCalendar:       onCalendar,
		SystemdProfile:   systemdProfile,
		Nice:             nice,
		Environment:      environment,
	}

	var data bytes.Buffer
	unitTmpl := template.Must(template.New("systemd.unit").Parse(systemdUnitBackupUnitTmpl))
	if err := unitTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName := filepath.Join(systemdUserDir, systemdProfile)
	clog.Infof("writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), defaultPermission); err != nil {
		return err
	}
	data.Reset()

	timerTmpl := template.Must(template.New("timer.unit").Parse(systemdUnitBackupTimerTmpl))
	if err := timerTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName = filepath.Join(systemdUserDir, timerProfile)
	clog.Infof("writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), defaultPermission); err != nil {
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
	if err := os.MkdirAll(systemdUserDir, 0700); err != nil {
		return "", err
	}
	return systemdUserDir, nil
}

// GetSystemDir returns the path where the local systemd units are stored
func GetSystemDir() string {
	return systemdSystemDir
}
