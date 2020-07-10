package systemd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/creativeprojects/resticprofile/clog"
)

const (
	defaultPermission = 0644
	systemdSystemDir  = "/etc/systemd/system/"

	systemdUnitBackupUnitTmpl = `[Unit]
Description={{ .JobDescription }}

[Service]
Type=oneshot
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .CommandLine }}
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
}

// Generate systemd unit
func Generate(commandLine, wd, title, subTitle, jobDescription, timerDescription string, onCalendar []string, unitType UnitType) error {
	var err error
	systemdProfile := GetServiceFile(title)
	timerProfile := GetTimerFile(title)

	systemdUserDir := systemdSystemDir
	if unitType == UserUnit {
		systemdUserDir, err = GetUserDir()
		if err != nil {
			return err
		}
	}

	info := TemplateInfo{
		JobDescription:   jobDescription,
		TimerDescription: timerDescription,
		WorkingDirectory: wd,
		CommandLine:      commandLine,
		OnCalendar:       onCalendar,
		SystemdProfile:   systemdProfile,
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
func GetServiceFile(profileName string) string {
	return fmt.Sprintf("resticprofile-backup@profile-%s.service", profileName)
}

// GetTimerFile returns the timer file name for the profile
func GetTimerFile(profileName string) string {
	return fmt.Sprintf("resticprofile-backup@profile-%s.timer", profileName)
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
