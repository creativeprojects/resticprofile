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
	// configurationExtensions list the possible extensions for the config file
	configurationExtensions = []string{
		"conf",
		"yaml",
		"toml",
		"json",
		"hcl",
	}

	defaultConfigurationLocationsUnix = []string{
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
		"c:\\restic\\",
		"c:\\resticprofile\\",
	}

	resticBinaryUnix    = "restic"
	resticBinaryWindows = "restic.exe"

	defaultBinaryLocationsUnix = []string{
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
// If the file doesn't have an extension, it will search for all possible extensions
func FindConfigurationFile(configFile string) (string, error) {
	found := ""
	extension := filepath.Ext(configFile)
	displayFile := ""
	if extension != "" {
		displayFile = fmt.Sprintf("'%s'", configFile)
		// Search only once through the paths
		found = findConfigurationFileWithExtension(configFile)
	} else {
		displayFile = fmt.Sprintf("'%s' with extensions %s", configFile, strings.Join(configurationExtensions, ", "))
		// Search all extensions one by one
		for _, ext := range configurationExtensions {
			found = findConfigurationFileWithExtension(configFile + "." + ext)
			if found != "" {
				break
			}
		}
	}
	if found != "" {
		return found, nil
	}
	// compile a list of search locations
	locations := append([]string{xdg.ConfigHome}, xdg.ConfigDirs...)
	locations = append(locations, getDefaultConfigurationLocations()...)
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, home)
	}
	return "", fmt.Errorf("configuration file %s was not found in the current directory nor any of these locations: %s", displayFile, strings.Join(locations, ", "))
}

func findConfigurationFileWithExtension(configFile string) string {
	// 1. Simple case: current folder (or rooted path)
	if fileExists(configFile) {
		return configFile
	}

	// 2. Next we try xdg as the "standard" for user configuration locations
	xdgFilename, err := xdg.SearchConfigFile(filepath.Join("resticprofile", configFile))
	if err == nil {
		if fileExists(xdgFilename) {
			return xdgFilename
		}
	}

	// 3. To keep compatibility with the older version in python, try the pre-selected locations
	paths := getDefaultConfigurationLocations()
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, home)
	}
	for _, configPath := range paths {
		filename := filepath.Join(configPath, configFile)
		if fileExists(filename) {
			return filename
		}
	}
	// Not found
	return ""
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
	return defaultConfigurationLocationsUnix
}

func getDefaultBinaryLocations() []string {
	if runtime.GOOS == "windows" {
		return defaultBinaryLocationsWindows
	}
	return defaultBinaryLocationsUnix
}

func getResticBinary() string {
	if runtime.GOOS == "windows" {
		return resticBinaryWindows
	}
	return resticBinaryUnix
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
