//go:build !darwin && !windows

package systemd

import (
	"bytes"
	"path"
	"strings"

	"github.com/spf13/afero"
)

func Read(unit string, unitType UnitType) (*Config, error) {
	var err error
	unitsDir := systemdSystemDir
	if unitType == UserUnit {
		unitsDir, err = GetUserDir()
		if err != nil {
			return nil, err
		}
	}
	content, err := afero.ReadFile(fs, path.Join(unitsDir, unit))
	if err != nil {
		return nil, err
	}
	currentSection := ""
	sections := make(map[string][]string, 3)
	lines := bytes.Split(content, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if bytes.HasPrefix(line, []byte("[")) && bytes.HasSuffix(line, []byte("]")) {
			// start of a section
			currentSection = string(bytes.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		if sections[currentSection] == nil {
			sections[currentSection] = []string{string(line)}
		} else {
			sections[currentSection] = append(sections[currentSection], string(line))
		}
	}
	unitSection, serviceSection := sections["Unit"], sections["Service"]
	description := getValue(unitSection, "Description")
	workdir := getValue(serviceSection, "WorkingDirectory")
	commandLine := getValue(serviceSection, "ExecStart")
	profileName, commandName := parseServiceFile(unit)
	cfg := &Config{
		Title:            profileName,
		SubTitle:         commandName,
		JobDescription:   description,
		WorkingDirectory: workdir,
		CommandLine:      commandLine,
		UnitType:         unitType,
	}
	return cfg, nil
}

func getValue(lines []string, key string) string {
	if len(lines) == 0 {
		return ""
	}
	for _, line := range lines {
		if k, v, found := strings.Cut(line, "="); found {
			k = strings.TrimSpace(k)
			if k == key {
				return strings.TrimSpace(v)
			}
		}
	}
	return ""
}

// parseServiceFile to detect profile and command names.
// format is: `resticprofile-backup@profile-name.service`
func parseServiceFile(filename string) (profileName, commandName string) {
	filename = strings.TrimPrefix(filename, "resticprofile-")
	filename = strings.TrimSuffix(filename, ".service")
	commandName, profileName, _ = strings.Cut(filename, "@")
	profileName = strings.TrimPrefix(profileName, "profile-")
	return
}
