package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendRemoteFiles(t *testing.T) {
	recorder := httptest.NewRecorder()
	sendRemoteFiles(&config.Remote{
		ConfigurationFile: "test_config.json",
		ProfileName:       "test_profile",
	}, "test_remote", []string{"arg1", "arg2"}, recorder)
	assert.Equal(t, recorder.Code, 200)
	assert.Equal(t, recorder.Header().Get("Content-Type"), "application/x-tar")
}

// getFreePort returns a TCP port number that is currently free on localhost.
func getFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
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
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("server did not become ready in time")
}

// stopServeCommand sends SIGINT to the current process, causing serveCommand's signal
// handler to trigger a graceful HTTP server shutdown.
func stopServeCommand(t *testing.T) {
	t.Helper()
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(os.Interrupt))
}

func TestServeCommandMissingPortArg(t *testing.T) {
	err := serveCommand(io.Discard, commandContext{Context: Context{
		flags: commandLineFlags{resticArgs: []string{"serve"}},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing argument: port")
}

func TestServeCommandHTTPRemoteNotFound(t *testing.T) {
	cfg, err := config.Load(bytes.NewBufferString("version: 2\n"), "yaml")
	require.NoError(t, err)

	port := getFreePort(t)
	cmd := commandContext{Context: Context{
		flags:  commandLineFlags{resticArgs: []string{"serve", fmt.Sprintf("%d", port)}},
		config: cfg,
	}}

	done := make(chan error, 1)
	go func() { done <- serveCommand(io.Discard, cmd) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	resp, err := http.Get(baseURL + "/configuration/nonexistent")
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	stopServeCommand(t)
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
`, remoteName)

	cfg, err := config.Load(bytes.NewBufferString(cfgYAML), "yaml")
	require.NoError(t, err)
	require.True(t, cfg.HasRemote(remoteName), "config should have the remote")

	port := getFreePort(t)
	cmd := commandContext{Context: Context{
		flags:  commandLineFlags{resticArgs: []string{"serve", fmt.Sprintf("%d", port)}},
		config: cfg,
	}}

	done := make(chan error, 1)
	go func() { done <- serveCommand(io.Discard, cmd) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	resp, err := http.Get(baseURL + "/configuration/" + remoteName)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/x-tar", resp.Header.Get("Content-Type"))

	stopServeCommand(t)
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
	cmd := commandContext{Context: Context{
		flags:  commandLineFlags{resticArgs: []string{"serve", fmt.Sprintf("%d", port)}},
		config: cfg,
	}}

	done := make(chan error, 1)
	go func() { done <- serveCommand(io.Discard, cmd) }()

	baseURL := fmt.Sprintf("http://localhost:%d", port)
	waitForServer(t, baseURL+"/configuration/probe")

	resp, err := http.Post(baseURL+"/configuration/myremote", "application/json", http.NoBody)
	require.NoError(t, err)
	resp.Body.Close()
	// The ServeMux pattern is method-specific; POST should not match
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	stopServeCommand(t)
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop in time")
	}
}
