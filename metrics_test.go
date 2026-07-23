package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMetricsHandlerMissingFile confirms that calling /metrics on a server
// whose textfile does not exist returns 503 Service Unavailable, with a
// message that names the missing path. Prometheus treats this as
// `up == 0` rather than logging a 404 every scrape.
func TestMetricsHandlerMissingFile(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "does-not-exist.prom")
	srv := httptestMetricsServer(t, textfile)
	resp := mustGetMetricsIdentity(t, srv.URL)
	body := mustReadBody(t, resp)
	resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	assert.Contains(t, body, "not found")
	assert.Contains(t, body, textfile)
}

// TestMetricsHandlerEmitsRuntimeAndTextfile confirms that, when the textfile
// exists, the response contains both the runtime collector output
// (`go_goroutines` is a stable signal we can grep for) and the body of the
// textfile appended verbatim after the runtime block.
func TestMetricsHandlerEmitsRuntimeAndTextfile(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "self.prom")
	require.NoError(t, os.WriteFile(textfile, []byte("# HELP x test counter\nx_total 7\n"), 0o644))

	srv := httptestMetricsServer(t, textfile)
	resp := mustGetMetricsIdentity(t, srv.URL)
	body := mustReadBody(t, resp)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, "go_goroutines", "runtime metrics should be present")
	assert.Contains(t, body, "x_total 7", "textfile body should be appended verbatim")
	// The textfile itself ends with a newline; we append it as-is after
	// promhttp's last line. The response must therefore end with the
	// textfile's final line.
	assert.True(t, strings.HasSuffix(body, "x_total 7\n"),
		"textfile body must be the last segment and end with a newline")
}

// TestMetricsHandlerTrailingNewlineGuard confirms the handler appends the
// textfile body with at least one separating newline, even when the file
// body itself does not end with a newline.
func TestMetricsHandlerTrailingNewlineGuard(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "no-trailing-newline.prom")
	require.NoError(t, os.WriteFile(textfile, []byte("missing_newline_at_end 1"), 0o644))

	srv := httptestMetricsServer(t, textfile)
	resp := mustGetMetricsIdentity(t, srv.URL)
	body := mustReadBody(t, resp)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, body, "go_goroutines")
	// The runtime collector always ends its body with a newline, so our
	// appended content is on a fresh line regardless of the file's tail.
	assert.Contains(t, body, "missing_newline_at_end 1")
	assert.True(t,
		strings.Contains(body, "\nmissing_newline_at_end 1") ||
			strings.HasSuffix(body, "\nmissing_newline_at_end 1\n"),
		"textfile line must start at a newline boundary, not glued to a prior partial line")
}

// TestMetricsHandlerGzipRoundTrip confirms that a scraper sending
// Accept-Encoding: gzip (Prometheus and Go's default client do) gets a single,
// valid gzip stream covering both the runtime block and the appended textfile.
// Before the fix promhttp gzipped only its own body and the raw textfile was
// appended past the end of that gzip member, so decoding failed with
// "gzip: invalid header".
func TestMetricsHandlerGzipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "self.prom")
	require.NoError(t, os.WriteFile(textfile, []byte("# HELP x test counter\nx_total 7\n"), 0o644))

	srv := httptestMetricsServer(t, textfile)

	// Disable the transport's transparent gzip so we decode the body ourselves
	// and would observe a corrupt stream if the response were malformed.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/metrics", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := (&http.Client{Transport: &http.Transport{DisableCompression: true}}).Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gz, err := gzip.NewReader(resp.Body)
	require.NoError(t, err, "response must be a single valid gzip stream")
	data, err := io.ReadAll(gz)
	require.NoError(t, err, "decoding the whole body must not hit invalid gzip data")
	require.NoError(t, gz.Close())

	body := string(data)
	assert.Contains(t, body, "go_goroutines", "runtime metrics should be present")
	assert.True(t, strings.HasSuffix(body, "x_total 7\n"),
		"textfile body must be the last segment inside the same gzip stream")
}

// TestMetricsServerGracefulShutdown starts the metrics server on a free port,
// scrapes /metrics once to make sure it's listening, and verifies that
// closing quit leads to a clean shutdown within the configured graceful window.
func TestMetricsServerGracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "shutdown.prom")
	require.NoError(t, os.WriteFile(textfile, []byte("# shutdown\n"), 0o644))

	srv := newMetricsServer(getFreePort(t), textfile)
	srv.timeouts.gracefulMax = 2 * time.Second // tighter for tests

	quit := make(chan os.Signal, 1)
	done := make(chan error, 1)
	go func() { done <- srv.run(quit) }()

	port := portFromAddr(t, srv.listen)
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForServer(t, baseURL+"/metrics")

	resp := mustGetMetricsIdentity(t, baseURL)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	quit <- os.Interrupt
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down in time after quit was closed")
	}
}

// TestMetricsServerPortInUse confirms that two servers racing for the same
// port produce a non-nil error from ListenAndServe (we surface it via .run).
// This is the realistic failure mode under `docker-compose scale`.
//
// Both servers bind on ":N" (all interfaces), so the second one must hit
// EADDRINUSE regardless of which interface the OS hands out.
func TestMetricsServerPortInUse(t *testing.T) {
	dir := t.TempDir()
	textfile := filepath.Join(dir, "occupied.prom")
	require.NoError(t, os.WriteFile(textfile, []byte("# x\n"), 0o644))

	port := getAllInterfacesFreePort(t)

	first := newMetricsServer(port, textfile)
	first.timeouts.gracefulMax = 1 * time.Second
	quit1 := make(chan os.Signal, 1)
	done1 := make(chan error, 1)
	go func() { done1 <- first.run(quit1) }()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	waitForServer(t, baseURL+"/metrics")

	// Second server on the same port must fail to bind.
	second := newMetricsServer(port, textfile)
	second.timeouts.gracefulMax = 1 * time.Second
	quit2 := make(chan os.Signal, 1)
	done2 := make(chan error, 1)
	go func() { done2 <- second.run(quit2) }()

	select {
	case err := <-done2:
		require.Error(t, err, "expected the second server to fail to bind")
		assert.NotContains(t, err.Error(), http.ErrServerClosed.Error())
	case <-time.After(3 * time.Second):
		t.Fatal("second server did not surface a bind error in time")
	}

	// Tear down the first server.
	close(quit1)
	select {
	case <-done1:
	case <-time.After(5 * time.Second):
		t.Fatal("first server did not shut down in time after quit was closed")
	}

	// second.run already returned its bind error above and is no longer
	// listening — nothing else to drain for it.
}

// getAllInterfacesFreePort returns a port number that is currently free on
// 0.0.0.0 (all interfaces), matching how the production metrics server binds.
func getAllInterfacesFreePort(t *testing.T) int {
	t.Helper()
	lc := net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", ":0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

// httptestMetricsServer wires a *metricsServer.handler behind httptest.Server
// on 127.0.0.1:<ephemeral>, so the tests above get a real HTTP roundtrip
// without depending on the package-level ListenAndServe path.
func httptestMetricsServer(t *testing.T, textfile string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", newMetricsServer(0, textfile).handler())
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// portFromAddr pulls the numeric port from an ":NNNN" or "host:NNNN" address.
// The metrics server always uses ":NNNN" (all interfaces), so the host part
// is irrelevant for tests.
func portFromAddr(t *testing.T, addr string) int {
	t.Helper()
	_, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)
	return port
}

// mustGetMetricsIdentity issues GET /metrics with Accept-Encoding: identity
// so the response body is plain text (promhttp auto-compresses otherwise and
// Go's default http.Client sets Accept-Encoding: gzip).
func mustGetMetricsIdentity(t *testing.T, baseURL string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/metrics", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// mustReadBody reads the entire HTTP response body and returns it as a string.
// Test helper — fatal on read errors so callers can keep the test body short.
func mustReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}
