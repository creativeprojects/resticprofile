package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTar creates a tar archive in memory from the given entries.
// Each entry is a map with keys "name" and "content".
func buildTar(t *testing.T, entries []struct{ name, content string }) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)
	for _, e := range entries {
		data := []byte(e.content)
		hdr := &tar.Header{
			Name: e.name,
			Size: int64(len(data)),
			Mode: 0644,
		}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	return buf.Bytes()
}

// newTarServer starts an httptest server that serves a tar archive with the given content-type.
func newTarServer(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-tar")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
}

func TestLoadRemoteFilesSuccess(t *testing.T) {
	manifest := remote.Manifest{
		Version:           "1.0",
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

	files, params, err := loadRemoteFiles(context.Background(), srv.URL)
	require.NoError(t, err)

	// manifest should be returned
	require.NotNil(t, params)
	assert.Equal(t, manifest.Version, params.Version)
	assert.Equal(t, manifest.ConfigurationFile, params.ConfigurationFile)
	assert.Equal(t, manifest.ProfileName, params.ProfileName)

	// one non-manifest file should be returned
	require.Len(t, files, 1)
}

func TestLoadRemoteFilesErrorNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, _, err := loadRemoteFiles(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "http error 404")
}

func TestLoadRemoteFilesWrongContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}))
	defer srv.Close()

	_, _, err := loadRemoteFiles(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected content type")
}

func TestLoadRemoteFilesInvalidEndpoint(t *testing.T) {
	// not a valid URL at all — NewRequestWithContext should fail
	_, _, err := loadRemoteFiles(context.Background(), "://invalid-url")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestLoadRemoteFilesUnreachableServer(t *testing.T) {
	_, _, err := loadRemoteFiles(context.Background(), "http://127.0.0.1:1") // nothing listening
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestLoadRemoteFilesPathTraversalRejected(t *testing.T) {
	entries := []struct{ name, content string }{
		{"../evil.sh", "rm -rf /"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	_, _, err := loadRemoteFiles(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid file name")
}

func TestLoadRemoteFilesCorruptTar(t *testing.T) {
	srv := newTarServer(t, []byte("this is not a tar archive"))
	defer srv.Close()

	_, _, err := loadRemoteFiles(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read tar header")
}

func TestLoadRemoteFilesInvalidManifestJSON(t *testing.T) {
	entries := []struct{ name, content string }{
		{constants.ManifestFilename, "not valid json {{{"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	_, _, err := loadRemoteFiles(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manifest")
}

func TestLoadRemoteFilesEmptyTar(t *testing.T) {
	tarBody := buildTar(t, nil)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	files, params, err := loadRemoteFiles(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Nil(t, params)
	assert.Empty(t, files)
}

func TestLoadRemoteFilesCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	srv := newTarServer(t, nil)
	defer srv.Close()

	_, _, err := loadRemoteFiles(ctx, srv.URL)
	require.Error(t, err)
}

func TestLoadMultipleRemoteFiles(t *testing.T) {
	manifest := remote.Manifest{ProfileName: "myprofile"}
	manifestJSON, err := json.Marshal(manifest)
	require.NoError(t, err)

	entries := []struct{ name, content string }{
		{constants.ManifestFilename, string(manifestJSON)},
		{"profiles.toml", "[profile]\n"},
		{"extra.conf", "key=value\n"},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	files, params, err := loadRemoteFiles(context.Background(), srv.URL)
	require.NoError(t, err)
	require.NotNil(t, params)
	assert.Equal(t, "myprofile", params.ProfileName)
	assert.Len(t, files, 2)
}

func TestLoadRemoteFilesWithBigFile(t *testing.T) {
	var size int64 = 10 * 1024 * 1024
	buffer := make([]byte, size) // 10 MB
	_, err := rand.NewChaCha8([32]byte{}).Read(buffer)
	require.NoError(t, err)

	entries := []struct{ name, content string }{
		{"bigfile", string(buffer)},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	files, _, err := loadRemoteFiles(ctx, srv.URL)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "bigfile", files[0].Name())
	assert.Equal(t, size, files[0].FileInfo().Size())
}

func TestLoadRemoteFilesWithEmptyFile(t *testing.T) {
	entries := []struct{ name, content string }{
		{"emptyfile", ""},
	}
	tarBody := buildTar(t, entries)

	srv := newTarServer(t, tarBody)
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	files, _, err := loadRemoteFiles(ctx, srv.URL)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "emptyfile", files[0].Name())
	assert.Equal(t, int64(0), files[0].FileInfo().Size())
}

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
	if os.Getenv("TEST_FUSE") == "" {
		t.Skip("set TEST_FUSE=1 to run FUSE-dependent tests")
	}

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
	if os.Getenv("TEST_FUSE") == "" {
		t.Skip("set TEST_FUSE=1 to run FUSE-dependent tests")
	}

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
