package main

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metricsServer exposes Go runtime metrics together with the Prometheus textfile
// written by a profile's prometheus-save-to-file on a single /metrics endpoint.
//
// Response contract:
//   - 200 OK with the runtime collector followed by the textfile body when the
//     file exists and is readable.
//   - 503 Service Unavailable when the file does not exist (e.g. the first
//     backup has not yet populated it). Prometheus treats 503 as a target
//     failure (up == 0) without flooding logs.
//   - 500 Internal Server Error for any other I/O error.
//
// Exposing the runtime collector and the textfile on the same /metrics
// endpoint matches what `node_exporter --collector.textfile` does, so the
// resulting file is also a drop-in for `prom.SaveTo`.
type metricsServer struct {
	textfile string
	listen   string
	timeouts metricsTimeouts
}

type metricsTimeouts struct {
	readHeader  time.Duration
	read        time.Duration
	write       time.Duration
	idle        time.Duration
	gracefulMax time.Duration
}

func defaultMetricsTimeouts() metricsTimeouts {
	return metricsTimeouts{
		readHeader:  5 * time.Second,
		read:        30 * time.Second,
		write:       30 * time.Second,
		idle:        120 * time.Second,
		gracefulMax: 30 * time.Second,
	}
}

// newMetricsServer returns a server bound on 0.0.0.0:port serving the
// textfile at textfile. The textfile path is operator-controlled via the
// `prometheus-save-to-file` profile setting.
func newMetricsServer(port int, textfile string) *metricsServer {
	return &metricsServer{
		textfile: textfile,
		listen:   fmt.Sprintf(":%d", port),
		timeouts: defaultMetricsTimeouts(),
	}
}

// handler returns the /metrics handler. Promhttp always emits valid
// exposition-format bodies terminated with a newline, so we can append
// the textfile body without a separator.
func (m *metricsServer) handler() http.Handler {
	// Disable promhttp's own gzip: it would compress only its runtime block,
	// and the raw textfile we append afterwards would land past the end of
	// that gzip member — every scraper then fails with "gzip: invalid header".
	// We compress the whole response (runtime + textfile) as one gzip stream
	// below instead.
	runtime := promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		DisableCompression: true,
	})
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		data, err := os.ReadFile(m.textfile) //nolint:gosec // textfile path is operator-supplied
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				clog.Debugf("metrics textfile %q not found yet", m.textfile)
				http.Error(resp, fmt.Sprintf("metrics file %q not found", m.textfile), http.StatusServiceUnavailable)
				return
			}
			http.Error(resp, fmt.Sprintf("reading metrics file: %v", err), http.StatusInternalServerError)
			return
		}

		// Let promhttp own the response headers, status and body prefix.
		// wrapWriter tees every body Write through ww.body — the gzip writer
		// when the client accepts it, the raw connection otherwise — so we can
		// append our textfile body afterwards without having to know whether
		// promhttp negotiated content-type or picked a non-200 status.
		ww := &wrapWriter{ResponseWriter: resp}
		if acceptsGzip(req) {
			resp.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(resp)
			defer gz.Close()
			ww.body = gz
		}
		runtime.ServeHTTP(ww, req)

		if ww.status >= 400 {
			// Runtime collector reported an error (or panic in user code).
			// Do not append the textfile body — the exposition format would
			// be invalid and Prometheus would log a parse error.
			return
		}

		if len(data) == 0 {
			return
		}
		if data[len(data)-1] != '\n' {
			if _, err := ww.Write([]byte{'\n'}); err != nil {
				clog.Debugf("failed to write newline before textfile body: %v", err)
				return
			}
		}
		if _, err := ww.Write(data); err != nil {
			clog.Debugf("failed to write textfile body: %v", err)
		}
	})
}

// wrapWriter is an http.ResponseWriter that forwards every call to the
// underlying writer while keeping track of the status code, so the handler
// can decide after-the-fact whether it is safe to append more bytes.
type wrapWriter struct {
	http.ResponseWriter
	body        io.Writer // body sink; when nil, writes go straight to ResponseWriter
	status      int
	wroteHeader bool
}

func (w *wrapWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return // ignore the second call, like net/http does
	}
	w.status = code
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *wrapWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		// Mirroring net/http default behaviour: a Write without an explicit
		// WriteHeader is treated as 200 OK.
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	if w.body != nil {
		return w.body.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

// Flush forwards to the underlying writer when it supports http.Flusher,
// which promhttp uses for streaming large metric sets. When gzipping, the
// gzip writer is flushed first so its buffered bytes reach the connection.
func (w *wrapWriter) Flush() {
	if gz, ok := w.body.(*gzip.Writer); ok {
		_ = gz.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// acceptsGzip reports whether the client offered gzip in Accept-Encoding.
// An absent header, a bare "identity" (as the tests send), or an explicit
// "gzip;q=0" all mean no.
func acceptsGzip(req *http.Request) bool {
	for _, part := range strings.Split(req.Header.Get("Accept-Encoding"), ",") {
		fields := strings.Split(strings.TrimSpace(part), ";")
		if !strings.EqualFold(strings.TrimSpace(fields[0]), "gzip") {
			continue
		}
		for _, p := range fields[1:] {
			if strings.EqualFold(strings.TrimSpace(p), "q=0") {
				return false
			}
		}
		return true
	}
	return false
}

// run starts the HTTP server and blocks until quit is closed. On shutdown it
// gives the server up to m.timeouts.gracefulMax to drain in-flight requests.
// Returns nil on clean shutdown (including http.ErrServerClosed), the bind
// or runtime error otherwise.
//
// The returned error can come from either ListenAndServe (e.g. EADDRINUSE) or
// from the shutdown goroutine; we surface whichever fires first so the caller
// can react without waiting for the process to exit.
func (m *metricsServer) run(quit <-chan os.Signal) error {
	server := &http.Server{
		Addr:              m.listen,
		Handler:           m.handler(),
		ReadHeaderTimeout: m.timeouts.readHeader,
		ReadTimeout:       m.timeouts.read,
		WriteTimeout:      m.timeouts.write,
		IdleTimeout:       m.timeouts.idle,
	}

	errCh := make(chan error, 2)

	go func() {
		<-quit
		clog.Info("shutting down the metrics server")
		ctx, cancel := context.WithTimeout(context.Background(), m.timeouts.gracefulMax)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			clog.Errorf("error while shutting down the metrics server: %v", err)
			errCh <- err
		}
	}()

	go func() {
		err := server.ListenAndServe()
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
		} else {
			errCh <- err
		}
	}()

	clog.Infof("metrics server listening on %s (textfile=%s)", server.Addr, m.textfile)
	if err := <-errCh; err != nil {
		return err
	}
	return nil
}
