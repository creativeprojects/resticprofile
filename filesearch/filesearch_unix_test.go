//go:build !windows

package filesearch

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindResticBinaryWithTilde(t *testing.T) {
	t.Parallel()

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

func TestAddRootToRelativePaths(t *testing.T) {
	t.Parallel()

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
