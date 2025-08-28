package filesearch

import (
	"fmt"
	iofs "io/fs"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/creativeprojects/resticprofile/platform"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	t.Parallel()

	t.Log("ConfigHome:", xdg.ConfigHome)
	t.Log("ConfigDirs:", xdg.ConfigDirs)
	t.Log("ApplicationDirs:", xdg.ApplicationDirs)
}

type testLocation struct {
	realPath   string
	realFile   string
	searchPath string
	searchFile string
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
			realPath:   "",
			realFile:   "profiles.yml",
			searchPath: "",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.spec",
			searchPath: "unittest-config",
			searchFile: "profiles.spec",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.conf",
			searchPath: "unittest-config",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.toml",
			searchPath: "unittest-config",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.yaml",
			searchPath: "unittest-config",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.json",
			searchPath: "unittest-config",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.hcl",
			searchPath: "unittest-config",
			searchFile: "profiles",
		},
		{
			realPath:   "unittest-config",
			realFile:   "profiles.yml",
			searchPath: "unittest-config",
			searchFile: "profiles",
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
		{
			realPath:   filepath.Join(xdg.ConfigHome, "resticprofile"),
			realFile:   "profiles.yml",
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
	t.Parallel()

	locations := testLocations(t)

	for _, location := range locations {
		t.Run(path.Join(location.realPath, location.realFile), func(t *testing.T) {
			t.Parallel()
			var err error
			fs := afero.NewMemMapFs()
			finder := Finder{fs: fs}

			// Install empty config file
			if location.realPath != "" {
				err = fs.MkdirAll(location.realPath, 0o700)
				require.NoError(t, err)
			}
			file, err := fs.Create(filepath.Join(location.realPath, location.realFile))
			require.NoError(t, err)
			require.NoError(t, file.Close())

			// Test
			found, err := finder.FindConfigurationFile(filepath.Join(location.searchPath, location.searchFile))
			assert.NoError(t, err)
			assert.NotEmpty(t, found)
			assert.Equal(t, filepath.Join(location.realPath, location.realFile), found)
		})
	}
}

func TestCannotFindConfigurationFile(t *testing.T) {
	t.Parallel()

	finder := NewFinder()
	found, err := finder.FindConfigurationFile("some_unknown-config_file")
	assert.Empty(t, found)
	assert.Error(t, err)
}

func TestFindResticBinary(t *testing.T) {
	t.Parallel()

	var paths []string
	if platform.IsWindows() {
		paths = defaultBinaryLocationsWindows
	} else {
		paths = defaultBinaryLocationsUnix
	}

	fs := afero.NewMemMapFs()
	file, err := fs.Create(path.Join(paths[len(paths)-1], getResticBinaryName()))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	finder := Finder{fs: fs}
	binary, err := finder.FindResticBinary("")

	assert.True(t, strings.HasSuffix(binary, getResticBinaryName()))
	assert.NoError(t, err)
}

func TestMayFindResticBinary(t *testing.T) {
	t.Parallel()

	fs := afero.NewMemMapFs()
	finder := Finder{fs: fs}
	binary, err := finder.FindResticBinary("")
	if binary != "" {
		// not found from fs, but latest resort is to search in the path
		assert.True(t, strings.HasSuffix(binary, getResticBinaryName()))
		assert.NoError(t, err)
	} else {
		assert.Error(t, err)
	}
}

func TestFindResticBinaryWithTilde(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("not supported on Windows")
		return
	}

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	fs := afero.NewMemMapFs()
	finder := Finder{fs: fs}

	tempFile, err := afero.TempFile(fs, home, t.Name())
	require.NoError(t, err)
	tempFile.Close()

	search := filepath.Join("~", filepath.Base(tempFile.Name()))
	binary, err := finder.FindResticBinary(search)
	require.NoError(t, err)
	assert.Equalf(t, tempFile.Name(), binary, "cannot find %q", search)
}

func TestShellExpand(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
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
			t.Parallel()

			result, err := ShellExpand(testItem.source)
			require.NoError(t, err)
			assert.Equal(t, testItem.expected, result)
		})
	}
}

func TestFindConfigurationIncludes(t *testing.T) {
	t.Parallel()

	testID := fmt.Sprintf("%x", time.Now().UnixNano())
	tempDir := os.TempDir()
	files := []string{
		filepath.Join(tempDir, "base."+testID+".conf"),
		filepath.Join(tempDir, "inc1."+testID+".conf"),
		filepath.Join(tempDir, "inc2."+testID+".conf"),
		filepath.Join(tempDir, "inc3."+testID+".conf"),
	}

	fs := afero.NewMemMapFs()
	for _, file := range files {
		require.NoError(t, afero.WriteFile(fs, file, []byte{}, iofs.ModePerm))
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

	finder := Finder{fs: fs}

	for _, test := range testData {
		t.Run(strings.Join(test.includes, ","), func(t *testing.T) {
			t.Parallel()

			result, err := finder.FindConfigurationIncludes(files[0], test.includes)
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

func TestAddRootToRelativePaths(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("not supported on Windows")
	}

	testCases := []struct {
		root       string
		inputPath  []string
		outputPath []string
	}{
		{
			root:       "",
			inputPath:  []string{"", "dir", "~/user", "/root"},
			outputPath: []string{"", "dir", "~/user", "/root"},
		},
		{
			root:       "/home",
			inputPath:  []string{"", "dir", "~/user", "/root"},
			outputPath: []string{"/home", "/home/dir", "/home/user", "/root"},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.root, func(t *testing.T) {
			t.Parallel()

			result := addRootToRelativePaths(testCase.root, testCase.inputPath)
			assert.Equal(t, testCase.outputPath, result)
		})
	}
}
