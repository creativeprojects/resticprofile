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
	path string
	file string
}

func TestFindConfigurationFileFromCurrentDirectory(t *testing.T) {
	// Work from a temporary directory
	err := os.Chdir(os.TempDir())
	require.NoError(t, err)
	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Log("Working directory:", cwd)

	locations := []testLocation{
		{"", "profiles.conf"},
		{"unittest-config", "profiles.conf"},
	}
	for _, location := range locations {
		var err error
		// Install empty config file
		if location.path != "" {
			err = os.MkdirAll(location.path, 0700)
			require.NoError(t, err)
		}
		file, err := os.Create(filepath.Join(location.path, location.file))
		require.NoError(t, err)
		file.Close()

		// Test
		found, err := FindConfigurationFile(filepath.Join(location.path, location.file))
		assert.NotEmpty(t, found)
		t.Log("Found", found)
		assert.NoError(t, err)

		// Clears up the test file
		if location.path == "" {
			err = os.Remove(location.file)
		} else {
			err = os.RemoveAll(location.path)
		}
		require.NoError(t, err)
	}
}
