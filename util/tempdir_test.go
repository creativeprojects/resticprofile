package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/util/shutdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTempDir(t *testing.T) {
	tempDirPrefix := tempDirPattern[:len(tempDirPattern)-1]

	t.Run("create-os.TempDir", func(t *testing.T) {
		dir, err := createTempDir("")
		defer removeTempDir(dir, err)

		assert.NoError(t, err)
		assert.Equal(t, filepath.Clean(os.TempDir()), filepath.Dir(dir))
		assert.True(t, strings.HasPrefix(filepath.Base(dir), tempDirPrefix))
		assert.NotEqual(t, filepath.Base(dir), tempDirPrefix)
	})

	t.Run("create-UserCacheDir-fallback", func(t *testing.T) {
		dir, err := createTempDir("non-existing-path")
		defer removeTempDir(dir, err)
		assert.NoError(t, err)

		cdir, err := os.UserCacheDir()
		require.NoError(t, err)

		assert.Equal(t, filepath.Clean(cdir), filepath.Dir(dir))
		assert.True(t, strings.HasPrefix(filepath.Base(dir), tempDirPrefix))
		assert.NotEqual(t, filepath.Base(dir), tempDirPrefix)
	})
}

func TestRemoveTempDir(t *testing.T) {
	dir, err := createTempDir("")
	defer removeTempDir(dir, err)

	assert.NoError(t, err)
	assert.DirExists(t, dir)

	removeTempDir(dir, err)
	assert.NoDirExists(t, dir)

	removeTempDir(dir, err)
	assert.NoDirExists(t, dir)

	removeTempDir("", err)
}

func TestTempDir(t *testing.T) {
	defer ClearTempDir()

	t.Run("stable-get", func(t *testing.T) {
		ClearTempDir()
		dir := MustGetTempDir()
		dir2 := MustGetTempDir()

		assert.DirExists(t, dir)
		assert.Equal(t, dir, dir2)
	})

	t.Run("can-clear", func(t *testing.T) {
		ClearTempDir()
		dir := MustGetTempDir()
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "test.file"), []byte("data"), 0700))

		ClearTempDir()
		dir2 := MustGetTempDir()

		assert.NotEqual(t, dir, dir2)
		assert.NoDirExists(t, dir)
		assert.DirExists(t, dir2)
	})

	t.Run("cleared-on-shutdown", func(t *testing.T) {
		ClearTempDir()
		dir := MustGetTempDir()
		assert.DirExists(t, dir)

		shutdown.RunHooks()
		assert.NoDirExists(t, dir)

		dir, err := TempDir()
		assert.Empty(t, dir)
		assert.ErrorContains(t, err, "illegal state: temp directory has been removed")

		ClearTempDir()
		dir = MustGetTempDir()
		assert.DirExists(t, dir)

		shutdown.RunHooks()
		assert.NoDirExists(t, dir)
	})
}
