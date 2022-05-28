package hook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestSendHook(t *testing.T) {

	testCases := []struct {
		cfg   config.SendMonitorSection
		calls int
	}{
		{config.SendMonitorSection{
			Method: http.MethodHead,
		}, 1},
		{config.SendMonitorSection{
			Method: http.MethodGet,
		}, 1},
		{config.SendMonitorSection{
			Method: http.MethodPost,
		}, 1},
		{config.SendMonitorSection{
			Method: http.MethodPost,
			Body:   "test body\n",
		}, 1},
		{config.SendMonitorSection{
			Method: http.MethodPost,
			Body:   "$PROFILE_NAME\n$PROFILE_COMMAND",
		}, 1},
		{config.SendMonitorSection{
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

			sender := NewSender("resticprofile_test", 10*time.Second)
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

	sender := NewSender("resticprofile_test", 300*time.Millisecond)
	err := sender.Send(config.SendMonitorSection{
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

	sender := NewSender("resticprofile_test", 300*time.Millisecond)
	// 1: request will fail TLS
	err := sender.Send(config.SendMonitorSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)
	assert.Equal(t, 0, calls)

	// 2: request allowing bad certificate
	err = sender.Send(config.SendMonitorSection{
		URL:     server.URL,
		SkipTLS: true,
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestFailedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewSender("resticprofile_test", 300*time.Millisecond)
	err := sender.Send(config.SendMonitorSection{
		URL: server.URL,
	}, Context{})
	assert.Error(t, err)
	t.Log(err)
}
