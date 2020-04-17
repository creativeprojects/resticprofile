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
Description=resticprofile backup for {{ .Profile }}

[Service]
Type=oneshot
ExecStart=resticprofile -n "{{ .Profile }}"
`

var systemdUnitBackupTimerTmpl = `[Unit]
Description=backup timer of {{ .Profile }}

[Timer]
OnCalendar={{.OnCalendar}}
Unit={{ .SystemdProfile }}
Persistent=true

[Install]
WantedBy=timers.target
`

type TemplateInfo struct {
	Profile        string
	OnCalendar     string
	SystemdProfile string
}

func Generate(profile, onCalendar string) error {
	systemdProfile := fmt.Sprintf("resticprofile-backup@%s.service", profile)
	timerProfile := fmt.Sprintf("resticprofile-backup@%s.timer", profile)

	u, err := user.Current()
	if err != nil {
		return err
	}

	systemdUserDir := filepath.Join(u.HomeDir, ".config", "systemd", "user")
	if err := os.MkdirAll(systemdUserDir, 0700); err != nil {
		return err
	}

	info := TemplateInfo{
		Profile:        profile,
		SystemdProfile: systemdProfile,
		OnCalendar:     onCalendar,
	}

	var data bytes.Buffer
	unitTmpl := template.Must(template.New("systemd.unit").Parse(systemdUnitBackupUnitTmpl))
	if err := unitTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName := filepath.Join(systemdUserDir, systemdProfile)
	clog.Infof("Writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), 0600); err != nil {
		return err
	}
	data.Reset()

	timerTmpl := template.Must(template.New("timer.unit").Parse(systemdUnitBackupTimerTmpl))
	if err := timerTmpl.Execute(&data, info); err != nil {
		return err
	}
	filePathName = filepath.Join(systemdUserDir, timerProfile)
	clog.Infof("Writing %v", filePathName)
	if err := ioutil.WriteFile(filePathName, data.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}
