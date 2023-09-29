package filesearch

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/xdg"
	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
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

var fs afero.Fs

func init() {
	fs = afero.NewOsFs()
}

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

	return "", fmt.Errorf(
		"configuration file %s was not found in the current directory nor any of these locations: %s",
		displayFile,
		strings.Join(getSearchConfigurationLocations(), ", "))
}

func findConfigurationFileWithExtension(configFile string) string {
	// simple case: try current folder (or rooted path)
	if fileExists(configFile) {
		return configFile
	}

	// try from a list of locations
	paths := getSearchConfigurationLocations()

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
		var err error
		if configFile, err = filepath.Abs(configFile); err != nil {
			return nil, err
		}
	}

	configFile = filepath.FromSlash(filepath.Clean(configFile))

	var files []string
	addFile := func(file string) {
		file = filepath.FromSlash(filepath.Clean(file))
		if file != configFile {
			clog.Tracef("include: %s", file)
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
			if matches, err := filepath.Glob(include); err == nil && matches != nil {
				sort.Strings(matches)
				for _, match := range matches {
					addFile(match)
				}
			} else if err == nil {
				clog.Debugf("no match: %s", include)
			} else {
				return nil, fmt.Errorf("%w: %q", err, include)
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
	paths := getSearchBinaryLocations()
	binaryFile := getResticBinaryName()

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
	if platform.IsWindows() {
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

func getSearchConfigurationLocations() []string {
	locations := []string{filepath.Join(xdg.ConfigHome, XDGAppName)}
	for _, configDir := range xdg.ConfigDirs {
		locations = append(locations, filepath.Join(configDir, XDGAppName))
	}

	if platform.IsWindows() {
		locations = append(locations, defaultConfigurationLocationsWindows...)
	} else {
		locations = append(locations, defaultConfigurationLocationsUnix...)
	}

	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, home)
	}

	// also adds binary dir on windows
	if platform.IsWindows() {
		if binary, err := os.Executable(); err == nil {
			locations = append(locations, filepath.Dir(binary))
		}
	}

	return locations
}

func getSearchBinaryLocations() []string {
	var paths []string
	if platform.IsWindows() {
		paths = defaultBinaryLocationsWindows
	} else {
		paths = defaultBinaryLocationsUnix
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".local/bin/"), filepath.Join(home, "bin/"))
	}
	return paths
}

func getResticBinaryName() string {
	if platform.IsWindows() {
		return resticBinaryWindows
	}
	return resticBinaryUnix
}

func fileExists(filename string) bool {
	_, err := fs.Stat(filename)
	return err == nil || errors.Is(err, iofs.ErrExist)
}
