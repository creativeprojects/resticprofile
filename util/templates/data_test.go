package templates

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBinaryDir(t *testing.T) {
	binary, err := os.Executable()
	require.NoError(t, err)
	binaryDir := filepath.ToSlash(filepath.Dir(binary))
	assert.Equal(t, binaryDir, NewDefaultData(nil).BinaryDir)
}

func TestCurrentDir(t *testing.T) {
	dir, err := os.Getwd()
	require.NoError(t, err)
	defer func(d string) { _ = os.Chdir(d) }(dir)

	require.NoError(t, os.Chdir(t.TempDir()))
	currentDir, _ := os.Getwd()

	assert.Equal(t, filepath.ToSlash(dir), NewDefaultData(nil).StartupDir)
	assert.Equal(t, filepath.ToSlash(currentDir), NewDefaultData(nil).CurrentDir)
}

func TestTempDir(t *testing.T) {
	assert.Equal(t, filepath.ToSlash(os.TempDir()), NewDefaultData(nil).TempDir)
}

func TestHostname(t *testing.T) {
	hostname, err := os.Hostname()
	require.NoError(t, err)
	assert.Equal(t, hostname, NewDefaultData(nil).Hostname)
}

func TestTime(t *testing.T) {
	assert.Equal(t, time.Now().Truncate(time.Second), NewDefaultData(nil).Now.Truncate(time.Second))
}

func TestEmptyInit(t *testing.T) {
	var data DefaultData

	now := time.Now().Truncate(time.Second)
	assert.NotEqual(t, now, data.Now.Truncate(time.Second))

	data.InitDefaults()
	assert.Equal(t, now, data.Now.Truncate(time.Second))
}

func TestOsAndArch(t *testing.T) {
	assert.Equal(t, runtime.GOOS, NewDefaultData(nil).OS)
	assert.Equal(t, runtime.GOARCH, NewDefaultData(nil).Arch)
}

func TestEnv(t *testing.T) {
	osEnv := util.NewDefaultEnvironment(os.Environ()...)

	customEnv := map[string]string{
		"path":      "my-test-path",
		"__test_k1": "v1",
		"__TEST_K2": "v2",
	}

	env := NewDefaultData(customEnv).Env

	for _, key := range osEnv.Names() {
		if key != "" && strings.ToUpper(key) != "PATH" {
			assert.Equal(t, os.Getenv(key), env[key], "key = %s", key)
		}
	}
	for key := range customEnv {
		rKey := osEnv.ResolveName(key)
		assert.Equal(t, customEnv[key], env[rKey], "key = %s, rKey = %s", key, rKey)
	}
}
