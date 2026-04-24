//go:build !windows && fuse

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRemoteConfigurationHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	closeFunc, manifest, err := setupRemoteConfiguration(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Nil(t, closeFunc)
	assert.Nil(t, manifest)
	assert.Contains(t, err.Error(), "http error 500")
}

func TestSetupRemoteConfigurationMissingManifest(t *testing.T) {
	entries := []struct{ name, content string }{
		{"profiles.toml", "[profile]\n"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	closeFunc, manifest, err := setupRemoteConfiguration(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Nil(t, closeFunc)
	assert.Nil(t, manifest)
	assert.Contains(t, err.Error(), "not found in remote configuration")
}

func TestSetupRemoteConfigurationCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	srv := newTarServer(t, nil)
	defer srv.Close()

	closeFunc, _, err := setupRemoteConfiguration(ctx, srv.URL)
	require.Error(t, err)
	assert.Nil(t, closeFunc)
}

func TestSetupRemoteConfigurationSuccess(t *testing.T) {
	manifest := remote.Manifest{
		ConfigurationFile: "profiles.toml",
		ProfileName:       "default",
	}
	manifestJSON, err := json.Marshal(manifest)
	require.NoError(t, err)

	entries := []struct{ name, content string }{
		{constants.ManifestFilename, string(manifestJSON)},
		{"profiles.toml", "[profile]\n"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	var closeFunc func()
	closeFunc, params, err := setupRemoteConfiguration(context.Background(), srv.URL)
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil && strings.Contains(err.Error(), "failed to mount filesystem") {
		t.Skipf("FUSE not available: %v", err)
	}
	require.NoError(t, err)
	require.NotNil(t, closeFunc)
	require.NotNil(t, params)
	assert.Equal(t, manifest.ConfigurationFile, params.ConfigurationFile)
	assert.Equal(t, manifest.ProfileName, params.ProfileName)

	// Working directory should now be the (temporary) mountpoint
	currentWd, err := os.Getwd()
	require.NoError(t, err)
	assert.NotEqual(t, originalWd, currentWd)

	// The config file should be accessible through the virtual FS
	_, statErr := os.Stat("profiles.toml")
	assert.NoError(t, statErr)

	// Call cleanup manually and verify the working directory is restored
	closeFunc()
	closeFunc = nil // prevent double-call from defer

	restoredWd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, originalWd, restoredWd)
}

func TestSetupRemoteConfigurationCustomMountpoint(t *testing.T) {
	mnt := t.TempDir()

	manifest := remote.Manifest{
		ConfigurationFile: "profiles.toml",
		ProfileName:       "default",
		Mountpoint:        mnt,
	}
	manifestJSON, err := json.Marshal(manifest)
	require.NoError(t, err)

	entries := []struct{ name, content string }{
		{constants.ManifestFilename, string(manifestJSON)},
		{"profiles.toml", "[profile]\n"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	originalWd, err := os.Getwd()
	require.NoError(t, err)

	var closeFunc func()
	closeFunc, params, err := setupRemoteConfiguration(context.Background(), srv.URL)
	defer func() {
		if closeFunc != nil {
			closeFunc()
		}
	}()
	if err != nil && strings.Contains(err.Error(), "failed to mount filesystem") {
		t.Skipf("FUSE not available: %v", err)
	}
	require.NoError(t, err)
	require.NotNil(t, closeFunc)
	require.NotNil(t, params)
	assert.Equal(t, mnt, params.Mountpoint)

	// Working directory should be the specified mountpoint
	currentWd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, mnt, currentWd)

	// Cleanup should restore the working directory but NOT remove the custom mountpoint
	closeFunc()
	closeFunc = nil

	restoredWd, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, originalWd, restoredWd)

	// The custom directory should still exist (cleanup doesn't own it)
	_, statErr := os.Stat(mnt)
	assert.NoError(t, statErr)
}
