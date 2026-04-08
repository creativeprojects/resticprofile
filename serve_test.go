package main

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRemoteFiles(t *testing.T) {
	recorder := httptest.NewRecorder()
	sendRemoteFiles(&config.Remote{
		ConfigurationFile: "examples/dev.yaml", // this file should exist in the test environment
		ProfileName:       "test_profile",
	}, "test_remote", []string{"arg1", "arg2"}, recorder)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "application/x-tar")
}

func TestSendRemoteFilesNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	sendRemoteFiles(&config.Remote{
		ConfigurationFile: "file-not-found", // this file should exist in the test environment
		ProfileName:       "test_profile",
	}, "test_remote", []string{"arg1", "arg2"}, recorder)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "text/plain")
	assert.True(t, strings.HasPrefix(recorder.Body.String(), "error while preparing files to send for remote \"test_remote\":"))
}

// getFreePort returns a TCP port number that is currently free on localhost.
func getFreePort(t *testing.T) int {
	t.Helper()
	listenConfig := net.ListenConfig{}
	l, err := listenConfig.Listen(context.Background(), "tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// waitForServer polls the given URL until the server responds or the deadline expires.
func waitForServer(t *testing.T, url string) {
	t.Helper()
	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
		require.NoError(t, err)
		resp, err := client.Do(request)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not become ready in time")
}

func TestServeCommandHTTPRemoteNotFound(t *testing.T) {
	cfg, err := config.Load(bytes.NewBufferString("version: 2\n"), "yaml")
	require.NoError(t, err)

	port := getFreePort(t)
	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- serveProfiles(strconv.Itoa(port), cfg, quit) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/configuration/nonexistent", http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	quit <- os.Interrupt
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServeCommandHTTPRemoteFound(t *testing.T) {
	const remoteName = "myremote"
	cfgYAML := fmt.Sprintf(`version: 2
remotes:
  %s:
    host: example.com:22
    username: testuser
    profile-name: default
    configuration-file: examples/dev.yaml
`, remoteName)

	cfg, err := config.Load(bytes.NewBufferString(cfgYAML), "yaml")
	require.NoError(t, err)
	require.True(t, cfg.HasRemote(remoteName), "config should have the remote")

	port := getFreePort(t)
	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- serveProfiles(strconv.Itoa(port), cfg, quit) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/configuration/"+remoteName, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/x-tar", resp.Header.Get("Content-Type"))

	quit <- os.Interrupt
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}

func TestServeCommandHTTPMethodNotAllowed(t *testing.T) {
	cfg, err := config.Load(bytes.NewBufferString("version: 2\n"), "yaml")
	require.NoError(t, err)

	port := getFreePort(t)
	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- serveProfiles(strconv.Itoa(port), cfg, quit) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/configuration/myremote", http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	resp.Body.Close()
	// The ServeMux pattern is method-specific; POST should not match
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	quit <- os.Interrupt
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}
