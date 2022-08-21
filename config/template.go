package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TemplateData contain the variables fed to a config template
type TemplateData struct {
	Profile    ProfileTemplateData
	Schedule   ScheduleTemplateData
	Now        time.Time
	CurrentDir string
	ConfigDir  string
	TempDir    string
	BinaryDir  string
	Hostname   string
	Env        map[string]string
}

// ProfileTemplateData contains profile data
type ProfileTemplateData struct {
	Name string
}

// ScheduleTemplateData contains schedule data
type ScheduleTemplateData struct {
	Name string
}

// newTemplateData populates a TemplateData struct ready to use
func newTemplateData(configFile, profileName, scheduleName string) TemplateData {
	currentDir, _ := os.Getwd()
	configDir := filepath.Dir(configFile)
	if !filepath.IsAbs(configDir) {
		configDir = filepath.Join(currentDir, configDir)
	}
	binary, _ := os.Executable()
	binaryDir := filepath.Dir(binary)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	env := make(map[string]string, len(os.Environ()))
	for _, envValue := range os.Environ() {
		keyValuePair := strings.SplitN(envValue, "=", 2)
		if keyValuePair[0] == "" {
			continue
		}
		env[keyValuePair[0]] = keyValuePair[1]
	}
	return TemplateData{
		Profile: ProfileTemplateData{
			Name: profileName,
		},
		Schedule: ScheduleTemplateData{
			Name: scheduleName,
		},
		Now:        time.Now(),
		ConfigDir:  configDir,
		CurrentDir: currentDir,
		TempDir:    os.TempDir(),
		BinaryDir:  binaryDir,
		Hostname:   hostname,
		Env:        env,
	}
}
