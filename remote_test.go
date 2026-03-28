package main

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
