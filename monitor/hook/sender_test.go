package hook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
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
