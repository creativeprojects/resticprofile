package filesearch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/creativeprojects/clog"
)

var (
	XDGAppName = "resticprofile"

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
	locations := []string{filepath.Join(xdg.ConfigHome, XDGAppName)}
	for _, configDir := range xdg.ConfigDirs {
		locations = append(locations, filepath.Join(configDir, XDGAppName))
	}
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
	xdgFilename, err := xdg.SearchConfigFile(filepath.Join(XDGAppName, configFile))
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

// FindConfigurationIncludes finds includes (glob patterns) relative to the configuration file.
func FindConfigurationIncludes(configFile string, includes []string) ([]string, error) {
	if !filepath.IsAbs(configFile) {
		if dir, err := os.Getwd(); err == nil {
			configFile = filepath.Join(dir, configFile)
		} else {
			return nil, err
		}
	}

	var files []string
	addFile := func(file string) {
		if file != configFile {
			files = append(files, file)
		}
	}

	base := filepath.Dir(configFile)
	for _, include := range includes {
		if !filepath.IsAbs(include) {
			include = filepath.Join(base, include)
		}

		if fileExists(include) {
			addFile(include)
		} else {
			if matches, err := filepath.Glob(include); err == nil {
				for _, match := range matches {
					addFile(match)
				}
			} else {
				return nil, err
			}
		}
	}

	return files, nil
}

// FindResticBinary returns the path of restic executable
func FindResticBinary(configLocation string) (string, error) {
	if configLocation != "" {
		// Start by the location from the configuration
		filename, err := ShellExpand(configLocation)
		if err != nil {
			clog.Warning(err)
		}
		if filename != "" && fileExists(filename) {
			return filename, nil
		}
		clog.Warningf("cannot find or read the restic binary specified in the configuration: %q", configLocation)
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

// ShellExpand uses the shell to expand variables and ~ from a filename.
// On Windows the function simply returns the filename unchanged
func ShellExpand(filename string) (string, error) {
	if runtime.GOOS == "windows" {
		return filename, nil
	}
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %s", strings.ReplaceAll(filename, " ", `\ `)))
	result, err := cmd.Output()
	if err != nil {
		return filename, err
	}
	filename = strings.TrimSuffix(string(result), "\n")
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
	return err == nil
}
