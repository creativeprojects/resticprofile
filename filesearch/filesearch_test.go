package filesearch

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Quick test to see the error message on the travis build agents:
//
// Linux:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: /home/travis/.config/some_service/some_path, /etc/xdg/some_service/some_path
//
// macOS:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: /Users/travis/Library/Preferences/some_service/some_path, /Library/Preferences/some_service/some_path
//
// Windows:
// filesearch_test.go:13: could not locate `some_file` in any of the following paths: C:\Users\travis\AppData\Local\some_service\some_path, C:\ProgramData\some_service\some_path
func TestSearchConfigFile(t *testing.T) {
	found, err := xdg.SearchConfigFile(filepath.Join("some_service", "some_path", "some_file"))
	t.Log(err)
	assert.Empty(t, found)
	assert.Error(t, err)
}

// Quick test to see the default xdg config on the build agents
//
// Linux:
// ConfigHome: /home/travis/.config
// ConfigDirs: [/etc/xdg]
//
// macOS:
// ConfigHome: /Users/travis/Library/Preferences
// ConfigDirs: [/Library/Preferences]
//
// Windows:
// ConfigHome: C:\Users\travis\AppData\Local
// ConfigDirs: [C:\ProgramData]
func TestDefaultConfigDirs(t *testing.T) {
	t.Log("ConfigHome:", xdg.ConfigHome)
	t.Log("ConfigDirs:", xdg.ConfigDirs)
}

type testLocation struct {
	realPath        string
	realFile        string
	searchPath      string
	searchFile      string
	deletePathAfter bool
}

func TestFindConfigurationFile(t *testing.T) {
	// Work from a temporary directory
	err := os.Chdir(os.TempDir())
	require.NoError(t, err)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Log("Working directory:", cwd)

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
	for _, location := range locations {
		var err error
		// Install empty config file
		if location.realPath != "" {
			err = os.MkdirAll(location.realPath, 0700)
			require.NoError(t, err)
		}
		file, err := os.Create(filepath.Join(location.realPath, location.realFile))
		require.NoError(t, err)
		file.Close()

		// Test
		found, err := FindConfigurationFile(filepath.Join(location.searchPath, location.searchFile))
		assert.NoError(t, err)
		assert.NotEmpty(t, found)
		assert.Equal(t, filepath.Join(location.realPath, location.realFile), found)

		// Clears up the test file
		if location.realPath == "" || !location.deletePathAfter {
			err = os.Remove(filepath.Join(location.realPath, location.realFile))
		} else {
			err = os.RemoveAll(location.realPath)
		}
		require.NoError(t, err)
	}
}
