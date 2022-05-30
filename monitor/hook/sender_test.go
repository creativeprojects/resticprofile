package hook

import (
	"bytes"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendHook(t *testing.T) {
	testCases := []struct {
		cfg   config.SendMonitoringSection
		calls int
	}{
		{config.SendMonitoringSection{
			Method: http.MethodHead,
		}, 1},
		{config.SendMonitoringSection{
			Method: http.MethodGet,
		}, 1},
		{config.SendMonitoringSection{
			Method: http.MethodPost,
		}, 1},
		{config.SendMonitoringSection{
			Method: http.MethodPost,
			Body:   "test body\n",
		}, 1},
		{config.SendMonitoringSection{
			Method: http.MethodPost,
			Body:   "$PROFILE_NAME\n$PROFILE_COMMAND",
		}, 1},
		{config.SendMonitoringSection{
			Method: http.MethodPost,
			Body:   "$ERROR\n$ERROR_COMMANDLINE\n$ERROR_EXIT_CODE\n$ERROR_STDERR\n",
		}, 1},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			calls := 0
			if testCase.cfg.URL == "" {
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, testCase.cfg.Method, r.Method)
					assert.Equal(t, "resticprofile_test", r.Header.Get("User-Agent"))
					assert.Equal(t, "/", r.URL.Path)
					buffer := bytes.Buffer{}
					_, err := buffer.ReadFrom(r.Body)
					assert.NoError(t, err)
					r.Body.Close()

					body := testCase.cfg.Body
					body = strings.ReplaceAll(body, "$PROFILE_NAME", "test_profile")
					body = strings.ReplaceAll(body, "$PROFILE_COMMAND", "test_command")
					body = strings.ReplaceAll(body, "$ERROR_COMMANDLINE", "test_command_line")
					body = strings.ReplaceAll(body, "$ERROR_EXIT_CODE", "test_exit_code")
					body = strings.ReplaceAll(body, "$ERROR_STDERR", "test_stderr")
					body = strings.ReplaceAll(body, "$ERROR", "test_error_message")
					assert.Equal(t, body, buffer.String())
					calls++
				})
				server := httptest.NewServer(handler)
				defer server.Close()
				testCase.cfg.URL = server.URL
			}

			ctx := Context{
				ProfileName:    "test_profile",
				ProfileCommand: "test_command",
				Error: ErrorContext{
					Message:     "test_error_message",
					CommandLine: "test_command_line",
					ExitCode:    "test_exit_code",
					Stderr:      "test_stderr",
				},
				Stdout: "test_stdout",
			}

			sender := NewSender(nil, "resticprofile_test", 10*time.Second)
			err := sender.Send(testCase.cfg, ctx)
			assert.NoError(t, err)

			assert.Equal(t, testCase.calls, calls)
		})
	}
}

func TestSenderTimeout(t *testing.T) {
	var startedCalls, finishedCalls uint32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&startedCalls, 1)
		time.Sleep(1 * time.Second)
		atomic.AddUint32(&finishedCalls, 1)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond)
	err := sender.Send(config.SendMonitoringSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)

	assert.Equal(t, uint32(1), atomic.LoadUint32(&startedCalls))
	assert.Equal(t, uint32(0), atomic.LoadUint32(&finishedCalls))
}

func TestInsecureRequests(t *testing.T) {
	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond)
	// 1: request will fail TLS
	err := sender.Send(config.SendMonitoringSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)
	assert.Equal(t, 0, calls)

	// 2: request allowing bad certificate
	err = sender.Send(config.SendMonitoringSection{
		URL:     server.URL,
		SkipTLS: true,
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRequestWithCA(t *testing.T) {
	defaultLogger := clog.GetDefaultLogger()
	clog.SetDefaultLogger(clog.NewLogger(clog.NewTestHandler(t)))
	defer clog.SetDefaultLogger(defaultLogger)

	calls := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond)
	// 1: request will fail TLS
	err := sender.Send(config.SendMonitoringSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)
	assert.Equal(t, 0, calls)

	// this is a bit hacky, but we need to save the certificate used by httptest
	assert.Equal(t, 1, len(server.TLS.Certificates))
	assert.Equal(t, 1, len(server.TLS.Certificates[0].Certificate))

	filename := filepath.Join(t.TempDir(), "ca.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]})
	err = os.WriteFile(filename, cert, 0o600)
	assert.NoError(t, err)

	// 2: request using the right CA certificate
	sender = NewSender([]string{filename}, "resticprofile_test", 300*time.Millisecond)
	err = sender.Send(config.SendMonitoringSection{
		URL: server.URL,
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestFailedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond)
	err := sender.Send(config.SendMonitoringSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)
}

func TestUserAgent(t *testing.T) {
	calls := 0
	testAgent := "test user agent/0.0"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testAgent, r.Header.Get("User-Agent"))
		calls++
	}))
	defer server.Close()

	sender := NewSender(nil, "", 300*time.Millisecond)
	err := sender.Send(config.SendMonitoringSection{
		URL: server.URL,
		Headers: []config.SendMonitoringHeader{
			{Name: "User-Agent", Value: testAgent},
		},
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestParseTemplate(t *testing.T) {
	ctx := Context{
		ProfileName: "test_profile",
	}

	template := `{{ .ProfileName }}-{{ .Error.ExitCode }}`
	filename := filepath.Join(t.TempDir(), "body.json")
	err := os.WriteFile(filename, []byte(template), 0o600)
	require.NoError(t, err)

	result, err := loadBodyTemplate(filename, ctx)
	assert.NoError(t, err)
	assert.Equal(t, "test_profile-", result)

	// test posting this body from template
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := &bytes.Buffer{}
		_, err = io.Copy(buffer, r.Body)
		assert.NoError(t, err)
		assert.Equal(t, result, buffer.String())
		r.Body.Close()
	}))
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond)
	err = sender.Send(config.SendMonitoringSection{
		URL:          server.URL,
		Method:       http.MethodPost,
		BodyTemplate: filename,
	}, ctx)
	assert.NoError(t, err)
}
