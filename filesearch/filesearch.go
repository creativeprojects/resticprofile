package filesearch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
)

var (
	defaultConfigurationLocationsPosix = []string{
		"./",
		"/usr/local/etc/",
		"/usr/local/etc/restic/",
		"/usr/local/etc/resticprofile/",
		"/etc/",
		"/etc/restic/",
		"/etc/resticprofile/",
		"/opt/local/etc/",
		"/opt/local/etc/restic/",
		"/opt/local/etc/resticprofile/",
	}

	defaultConfigurationLocationsWindows = []string{
		".\\",
		"c:\\restic\\",
		"c:\\resticprofile\\",
	}

	resticBinaryPosix   = "restic"
	resticBinaryWindows = "restic.exe"

	defaultBinaryLocationsPosix = []string{
		"/usr/bin/",
		"/usr/local/bin/",
		"./",
		"/opt/local/bin/",
	}

	defaultBinaryLocationsWindows = []string{
		"c:\\ProgramData\\chocolatey\\bin\\",
		"c:\\restic\\",
		"c:\\resticprofile\\",
		"c:\\tools\\restic\\",
		"c:\\tools\\resticprofile\\",
		".\\",
	}
)

// FindConfigurationFile returns the path of the configuration file
func FindConfigurationFile(configFile string) (string, error) {
	xdgFilename, err := xdg.ConfigFile(filepath.Join("resticprofile", configFile))
	if err == nil {
		if fileExists(xdgFilename) {
			return xdgFilename, nil
		}
	}

	paths := getDefaultConfigurationLocations()
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, home)
	}
	for _, configPath := range paths {
		filename := filepath.Join(configPath, configFile)
		if fileExists(filename) {
			return filename, nil
		}
	}
	return "", fmt.Errorf("Configuration file '%s' was not found in any of these locations: %s", configFile, strings.Join(paths, ", "))
}

// FindResticBinary returns the path of restic executable
func FindResticBinary(configLocation string) (string, error) {
	// Start by the location from the configuration
	if configLocation != "" && fileExists(configLocation) {
		return configLocation, nil
	}
	paths := getDefaultBinaryLocations()
	binaryFile := getResticBinary()
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".local/bin/"))
	}
	for _, configPath := range paths {
		filename := filepath.Join(configPath, binaryFile)
		if fileExists(filename) {
			return filename, nil
		}
	}
	// Last resort, search from the OS PATH
	filename, err := exec.LookPath(binaryFile)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func getDefaultConfigurationLocations() []string {
	if runtime.GOOS == "windows" {
		return defaultConfigurationLocationsWindows
	}
	return defaultConfigurationLocationsPosix
}

func getDefaultBinaryLocations() []string {
	if runtime.GOOS == "windows" {
		return defaultBinaryLocationsWindows
	}
	return defaultBinaryLocationsPosix
}

func getResticBinary() string {
	if runtime.GOOS == "windows" {
		return resticBinaryWindows
	}
	return resticBinaryPosix
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
