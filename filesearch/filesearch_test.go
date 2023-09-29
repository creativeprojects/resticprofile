package filesearch

import (
	"fmt"
	iofs "io/fs"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.M) {
	fs = afero.NewMemMapFs()
	os.Exit(t.Run())
}

// Quick test to see the default xdg config on the build agents
//
// Linux:
// ConfigHome: /home/runner/.config
// ConfigDirs: [/etc/xdg]
// ApplicationDirs: [/home/runner/.local/share/applications /usr/local/share/applications /usr/share/applications]
//
// macOS:
// ConfigHome: /Users/runner/Library/Application Support
// ConfigDirs: [/Users/runner/Library/Preferences /Library/Application Support /Library/Preferences]
// ApplicationDirs: [/Applications]
//
// Windows:
// ConfigHome: C:\Users\runneradmin\AppData\Local
// ConfigDirs: [C:\ProgramData C:\Users\runneradmin\AppData\Roaming]
// ApplicationDirs: [C:\Users\runneradmin\AppData\Roaming\Microsoft\Windows\Start Menu\Programs C:\ProgramData\Microsoft\Windows\Start Menu\Programs]
func TestDefaultConfigDirs(t *testing.T) {
	t.Log("ConfigHome:", xdg.ConfigHome)
	t.Log("ConfigDirs:", xdg.ConfigDirs)
	t.Log("ApplicationDirs:", xdg.ApplicationDirs)
}

type testLocation struct {
	realPath        string
	realFile        string
	searchPath      string
	searchFile      string
	deletePathAfter bool
}

func testLocations(t *testing.T) []testLocation {
	t.Helper()

	binary, err := os.Executable()
	require.NoError(t, err)
	binaryDir := filepath.Dir(binary)
	t.Log("Binary directory:", binaryDir)

	locations := []testLocation{
		{
			realPath:   "",
			realFile:   "profiles.spec",
			searchPath: "",
			searchFile: "profiles.spec",
		},
		{
			realPath:   "",
			realFile:   "profiles.conf",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.yaml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.json",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.toml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "",
			realFile:   "profiles.hcl",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.spec",
			searchPath:      "unittest-config",
			searchFile:      "profiles.spec",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.conf",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.toml",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.yaml",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.json",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:        "unittest-config",
			realFile:        "profiles.hcl",
			searchPath:      "unittest-config",
			searchFile:      "profiles",
			deletePathAfter: true,
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.spec",
			searchPath: "",
			searchFile: "profiles.spec",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.conf",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.toml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.yaml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.json",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.hcl",
			searchPath: "",
			searchFile: "profiles",
		},
	}

	// on windows, we allow config file next to the resticprofile executable
	if platform.IsWindows() {
		locations = append(locations, testLocation{
			realPath:   binaryDir,
			realFile:   "profiles.conf",
			searchPath: "",
			searchFile: "profiles",
		})
	}

	return locations
}

func TestFindConfigurationFile(t *testing.T) {
	// Work from a temporary directory
	err := os.Chdir(os.TempDir())
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Log("Working directory:", cwd)

	locations := testLocations(t)

	for _, location := range locations {
		var err error
		// Install empty config file
		if location.realPath != "" {
			err = fs.MkdirAll(location.realPath, 0700)
			require.NoError(t, err)
		}
		file, err := fs.Create(filepath.Join(location.realPath, location.realFile))
		require.NoError(t, err)
		file.Close()

		// Test
		found, err := FindConfigurationFile(filepath.Join(location.searchPath, location.searchFile))
		assert.NoError(t, err)
		assert.NotEmpty(t, found)
		assert.Equal(t, filepath.Join(location.realPath, location.realFile), found)

		// Clears up the test file
		if location.realPath == "" || !location.deletePathAfter {
			err = fs.Remove(filepath.Join(location.realPath, location.realFile))
		} else {
			err = fs.RemoveAll(location.realPath)
		}
		require.NoError(t, err)
	}
}

func TestCannotFindConfigurationFile(t *testing.T) {
	found, err := FindConfigurationFile("some_config_file")
	assert.Empty(t, found)
	assert.Error(t, err)
}

func TestFindResticBinary(t *testing.T) {
	binary, err := FindResticBinary("some_other_name")
	if binary != "" {
		assert.True(t, strings.HasSuffix(binary, getResticBinaryName()))
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}

func TestFindResticBinaryWithTilde(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on Windows")
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tempFile, err := afero.TempFile(fs, home, "TestFindResticBinaryWithTilde")
	require.NoError(t, err)
	tempFile.Close()
	defer func() {
		fs.Remove(tempFile.Name())
	}()

	search := filepath.Join("~", filepath.Base(tempFile.Name()))
	binary, err := FindResticBinary(search)
	require.NoError(t, err)
	assert.Equalf(t, tempFile.Name(), binary, "cannot find %q", search)
}

func TestShellExpand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on Windows")
		return
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	usr, err := user.Current()
	require.NoError(t, err)

	testData := []struct {
		source   string
		expected string
	}{
		{"/", "/"},
		{"~", home},
		{"$HOME", home},
		{"~" + usr.Username, usr.HomeDir},
		{"1 2", "1 2"},
	}

	for _, testItem := range testData {
		t.Run(testItem.source, func(t *testing.T) {
			result, err := ShellExpand(testItem.source)
			require.NoError(t, err)
			assert.Equal(t, testItem.expected, result)
		})
	}
}

func TestFindConfigurationIncludes(t *testing.T) {
	testID := fmt.Sprintf("%d", uint32(time.Now().UnixNano()))
	tempDir := os.TempDir()
	files := []string{
		filepath.Join(tempDir, "base."+testID+".conf"),
		filepath.Join(tempDir, "inc1."+testID+".conf"),
		filepath.Join(tempDir, "inc2."+testID+".conf"),
		filepath.Join(tempDir, "inc3."+testID+".conf"),
	}

	for _, file := range files {
		require.NoError(t, afero.WriteFile(fs, file, []byte{}, iofs.ModePerm))
		defer fs.Remove(file) // defer stack is ok for cleanup
	}

	testData := []struct {
		includes []string
		expected []string
	}{
		// Invalid pattern
		{[]string{"[--]"}, nil},
		// Empty
		{[]string{"no-match"}, []string{}},
		// Existing files
		{files[2:4], files[2:4]},
		// GLOB patterns
		{[]string{"inc*." + testID + ".conf"}, files[1:]},
		{[]string{"*inc*." + testID + ".*"}, files[1:]},
		{[]string{"inc1." + testID + ".conf"}, files[1:2]},
		{[]string{"inc3." + testID + ".conf", "inc1." + testID + ".conf"}, []string{files[3], files[1]}},
		{[]string{"inc3." + testID + ".conf", "no-match"}, []string{files[3]}},
		// Does not include self
		{[]string{"base." + testID + ".conf"}, []string{}},
		{files[0:1], []string{}},
	}

	for _, test := range testData {
		t.Run(strings.Join(test.includes, ","), func(t *testing.T) {
			result, err := FindConfigurationIncludes(files[0], test.includes)
			if test.expected == nil {
				assert.Nil(t, result)
				assert.NotNil(t, err)
			} else {
				assert.NoError(t, err)
				if len(test.expected) == 0 {
					assert.Nil(t, result)
				} else {
					assert.Equal(t, test.expected, result)
				}
			}
		})
	}
}
