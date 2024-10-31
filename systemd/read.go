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
	sections := make(map[string]map[string][]string, 3)
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
		if key, value, found := strings.Cut(string(line), "="); found {
			value = strings.Trim(value, `"`)
			if sections[currentSection] == nil {
				sections[currentSection] = map[string][]string{
					key: {value},
				}
			} else if sections[currentSection][key] == nil {
				sections[currentSection][key] = []string{value}
			} else {
				sections[currentSection][key] = append(sections[currentSection][key], value)
			}
		}
	}
	description := getSingleValue(sections, "Unit", "Description")
	workdir := getSingleValue(sections, "Service", "WorkingDirectory")
	commandLine := getSingleValue(sections, "Service", "ExecStart")
	environment := getValues(sections, "Service", "Environment")
	profileName, commandName := parseServiceFileName(unit)
	cfg := &Config{
		Title:            profileName,
		SubTitle:         commandName,
		JobDescription:   description,
		WorkingDirectory: workdir,
		CommandLine:      commandLine,
		UnitType:         unitType,
		Environment:      environment,
	}
	return cfg, nil
}

func getSingleValue(from map[string]map[string][]string, section, key string) string {
	if section, found := from[section]; found {
		if values, found := section[key]; found {
			if len(values) > 0 {
				return values[0]
			}
		}
	}
	return ""
}

func getValues(from map[string]map[string][]string, section, key string) []string {
	if section, found := from[section]; found {
		if values, found := section[key]; found {
			return values
		}
	}
	return nil
}

// parseServiceFileName to detect profile and command names from the file name.
// format is: `resticprofile-backup@profile-name.service`
func parseServiceFileName(filename string) (profileName, commandName string) {
	filename = strings.TrimPrefix(filename, "resticprofile-")
	filename = strings.TrimSuffix(filename, ".service")
	commandName, profileName, _ = strings.Cut(filename, "@")
	profileName = strings.TrimPrefix(profileName, "profile-")
	return
}
