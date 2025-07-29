package hook

import (
	"bytes"
	"encoding/pem"
	"fmt"
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
	"github.com/creativeprojects/resticprofile/constants"
)

func TestSend(t *testing.T) {
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
			Body:   "test $$escaped\n",
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

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			calls := 0
			if testCase.cfg.URL.Value() == "" {
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
					body = strings.ReplaceAll(body, "$$", "$")
					assert.Equal(t, body, buffer.String())
					calls++
				})
				server := httptest.NewServer(handler)
				defer server.Close()
				testCase.cfg.URL = config.NewConfidentialValue(server.URL)
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

			sender := NewSender(nil, "resticprofile_test", 10*time.Second, false)
			err := sender.Send(testCase.cfg, ctx)
			assert.NoError(t, err)

			assert.Equal(t, testCase.calls, calls)
		})
	}
}

func TestDryRun(t *testing.T) {
	var calls uint32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint32(&calls, 1)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	sender := NewSender(nil, "", time.Second, true)
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
	}, Context{})
	assert.NoError(t, err)

	assert.Equal(t, uint32(0), atomic.LoadUint32(&calls))
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

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond, false)
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
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

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond, false)
	// 1: request will fail TLS
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
	}, Context{})
	assert.Error(t, err)
	assert.Equal(t, 0, calls)

	// 2: request allowing bad certificate
	err = sender.Send(config.SendMonitoringSection{
		URL:     config.NewConfidentialValue(server.URL),
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

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond, false)
	// 1: request will fail TLS
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
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
	sender = NewSender([]string{filename}, "resticprofile_test", 300*time.Millisecond, false)
	err = sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestFailedRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond, false)
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
	}, Context{})
	assert.Error(t, err)
}

func TestUserAgent(t *testing.T) {
	calls := 0
	agentHeader := "User-Agent"
	testAgent := "test user agent/0.0"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, testAgent, r.Header.Get(agentHeader))
		calls++
	}))
	defer server.Close()

	sender := NewSender(nil, "", 300*time.Millisecond, false)
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(server.URL),
		Headers: []config.SendMonitoringHeader{
			{Name: agentHeader, Value: config.NewConfidentialValue(testAgent)},
		},
	}, Context{})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestConfidentialURL(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "user", user)
		assert.Equal(t, "password", password)
		calls++
	}))
	defer server.Close()

	serverURL := strings.Replace(server.URL, "http://", "http://user:password@", 1)

	profile := &config.Profile{
		Backup: &config.BackupSection{
			GenericSectionWithSchedule: config.GenericSectionWithSchedule{
				GenericSection: config.GenericSection{
					SendMonitoringSections: config.SendMonitoringSections{
						SendBefore: []config.SendMonitoringSection{
							{
								URL: config.NewConfidentialValue(serverURL),
							},
						},
					},
				},
			},
		},
	}
	config.ProcessConfidentialValues(profile)

	sender := NewSender(nil, "", 300*time.Millisecond, false)
	err := sender.Send(profile.Backup.SendBefore[0], Context{})
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestURLEncoding(t *testing.T) {
	ctx := Context{
		ProfileName:    "unencoded/name",
		ProfileCommand: "unencoded/command",
		Error: ErrorContext{
			Message:     "some/error/message",
			CommandLine: "some < tricky || command & line",
			ExitCode:    "1",
			Stderr:      "some\nmultiline\nerror\nwith strange &/~!^., characters",
		},
		Stdout: "unused",
	}

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		assert.Equal(t, fmt.Sprintf("/%s-%s", ctx.ProfileName, ctx.ProfileCommand), r.URL.Path)

		assert.Equal(t, ctx.Error.Message, query.Get("message"))
		assert.Equal(t, ctx.Error.CommandLine, query.Get("command_line"))
		assert.Equal(t, ctx.Error.ExitCode, query.Get("exit_code"))
		assert.Equal(t, ctx.Error.Stderr, query.Get("stderr"))

		assert.Equal(t, "$TEST_MONITOR_URL", query.Get("escaped"))

		calls++
	}))
	defer server.Close()

	// test if env vars are untouched
	t.Setenv("TEST_MONITOR_URL", server.URL)

	serverURL := fmt.Sprintf(
		"$TEST_MONITOR_URL/$%s-$%s?message=$%s&command_line=$%s&exit_code=$%s&stderr=$%s&escaped=$$TEST_MONITOR_URL",
		constants.EnvProfileName,
		constants.EnvProfileCommand,
		constants.EnvError,
		constants.EnvErrorCommandLine,
		constants.EnvErrorExitCode,
		constants.EnvErrorStderr,
	)

	sender := NewSender(nil, "", 300*time.Millisecond, false)
	err := sender.Send(config.SendMonitoringSection{
		URL: config.NewConfidentialValue(serverURL),
	}, ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestConfidentialHeader(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	calls := 0
	headerKey := "Authorization"
	headerValue := "Bearer secret_token"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, headerValue, r.Header.Get(headerKey))
		calls++
	}))
	defer server.Close()

	profile := &config.Profile{
		Backup: &config.BackupSection{
			GenericSectionWithSchedule: config.GenericSectionWithSchedule{
				GenericSection: config.GenericSection{
					SendMonitoringSections: config.SendMonitoringSections{
						SendBefore: []config.SendMonitoringSection{
							{
								URL: config.NewConfidentialValue(server.URL),
								Headers: []config.SendMonitoringHeader{
									{Name: headerKey, Value: config.NewConfidentialValue(headerValue)},
								},
							},
						},
					},
				},
			},
		},
	}
	config.ProcessConfidentialValues(profile)

	sender := NewSender(nil, "", 300*time.Millisecond, false)
	err := sender.Send(profile.Backup.SendBefore[0], Context{})
	require.NoError(t, err)
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

	sender := NewSender(nil, "resticprofile_test", 300*time.Millisecond, false)
	err = sender.Send(config.SendMonitoringSection{
		URL:          config.NewConfidentialValue(server.URL),
		Method:       http.MethodPost,
		BodyTemplate: filename,
	}, ctx)
	assert.NoError(t, err)
}

func TestResponseSanitizer(t *testing.T) {
	var tests [][]string
	for i := 0; i < 256; i++ {
		if (i < 32 && i != '\f' && i != '\t' && i != '\r' && i != '\n') || i > 127 {
			tests = append(tests, []string{string([]byte{byte(i)}), " "})
		}
	}

	tests = append(tests, [][]string{
		{`[{"key": "value"}]`, `[{"key": "value"}]`},
		{`<x a="2"></x>`, `<x a="2"></x>`},
		{`{{((_-'987654321'-_*,.;:!"§$%&"))}}`, `{{((_-'987654321'-_*,.;:!"§$%&"))}}`},
		{"\r\n\t ", "\r\n\t "},
		{"\r\n", "\r\n"},
		{"\ufeffxyc", " xyc"},
	}...)

	for i, test := range tests {
		assert.Equal(t, test[1], responseContentSanitizer.ReplaceAllString(test[0], " "), "test #%d", i)
	}
}
