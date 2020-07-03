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

var systemdUnitBackupUnitTmpl = `[Unit]
Description=resticprofile backup for profile '{{ .ProfileName }}'

[Service]
Type=oneshot
WorkingDirectory={{ .WorkingDirectory }}
ExecStart={{ .Binary }} --no-ansi --config "{{ .ConfigFile }}" --name "{{ .ProfileName }}" backup
`

var systemdUnitBackupTimerTmpl = `[Unit]
Description=backup timer for profile '{{ .ProfileName }}'

[Timer]
{{ range .OnCalendar }}OnCalendar={{ . }}{{ end }}
Unit={{ .SystemdProfile }}
Persistent=true

[Install]
WantedBy=timers.target
`

// TemplateInfo to create systemd unit
type TemplateInfo struct {
	WorkingDirectory string
	Binary           string
	ConfigFile       string
	ProfileName      string
	OnCalendar       []string
	SystemdProfile   string
}

// Generate systemd unit
func Generate(wd, binary, configFile, profileName string, onCalendar []string) error {
	systemdProfile := GetServiceFile(profileName)
	timerProfile := GetTimerFile(profileName)

	systemdUserDir, err := GetUserDir()
	if err != nil {
		return err
	}

	info := TemplateInfo{
		WorkingDirectory: wd,
		Binary:           binary,
		ConfigFile:       configFile,
		ProfileName:      profileName,
		SystemdProfile:   systemdProfile,
		OnCalendar:       onCalendar,
	}

	var data bytes.Buffer
	unitTmpl := template.Must(template.New("systemd.unit").Parse(systemdUnitBackupUnitTmpl))
	if err := unitTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName := filepath.Join(systemdUserDir, systemdProfile)
	clog.Infof("writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), 0600); err != nil {
		return err
	}
	data.Reset()

	timerTmpl := template.Must(template.New("timer.unit").Parse(systemdUnitBackupTimerTmpl))
	if err := timerTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName = filepath.Join(systemdUserDir, timerProfile)
	clog.Infof("writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), 0600); err != nil {
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
