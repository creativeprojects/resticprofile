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
	Now        time.Time
	CurrentDir string
	ConfigDir  string
	TempDir    string
	Hostname   string
	Env        map[string]string
}

// ProfileTemplateData contains profile data
type ProfileTemplateData struct {
	Name string
}

// newTemplateData populates a TemplateData struct ready to use
func newTemplateData(configFile, profileName string) TemplateData {
	currentDir, _ := os.Getwd()
	configDir := filepath.Dir(configFile)
	if !filepath.IsAbs(configDir) {
		configDir = filepath.Join(currentDir, configDir)
	}
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
		Now:        time.Now(),
		ConfigDir:  configDir,
		CurrentDir: currentDir,
		TempDir:    os.TempDir(),
		Hostname:   hostname,
		Env:        env,
	}
}
