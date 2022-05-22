package hc

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandProgressStart(t *testing.T) {
	fixtures := []struct {
		command       string
		uuid          string
		expectedCalls int
	}{
		{constants.CommandBackup, "test-uuid", 1},
		{constants.CommandBackup, "", 0},
		{"other", "test-uuid", 0},
	}

	for _, test := range fixtures {
		t.Run(test.command+"-"+test.uuid, func(t *testing.T) {
			calls := 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodHead, r.Method)
				assert.Equal(t, "/"+test.uuid+"/start", r.URL.Path)
				calls++
			})
			server := httptest.NewServer(handler)

			healthchecks := NewProgress(&config.Profile{
				HealthChecksURL: server.URL,
				Backup: &config.BackupSection{
					HealthChecksUUID: test.uuid,
				},
			})

			healthchecks.Start(test.command)
			assert.Equal(t, test.expectedCalls, calls)
		})
	}
}

func TestProgressStartTimeout(t *testing.T) {

	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		calls++
	})
	server := httptest.NewServer(handler)

	healthchecks := NewProgress(&config.Profile{
		HealthChecksURL:     server.URL,
		HealthChecksTimeout: 300 * time.Millisecond,
		Backup: &config.BackupSection{
			HealthChecksUUID: "test-uuid",
		},
	})

	healthchecks.Start(constants.CommandBackup)
	assert.Equal(t, 0, calls)
}

func TestCommandProgressSummary(t *testing.T) {
	type testStruct struct {
		command       string
		uuid          string
		stderr        string
		success       bool
		expectedCalls int
	}
	fixtures := []testStruct{
		{"other", "test-uuid", "", true, 0},
		{"other", "test-uuid", "", false, 0},
	}

	commands := []string{
		constants.CommandBackup,
		// constants.CommandCheck,
		// constants.CommandCopy,
		// constants.CommandForget,
		// constants.CommandPrune,
	}
	for _, command := range commands {
		fixtures = append(fixtures, []testStruct{
			{command, "test-uuid", "", true, 1},
			{command, "test-uuid", "some warnings in stderr", true, 1},
			{command, "test-uuid", "", false, 1},
			{command, "test-uuid", "some warnings in stderr", false, 1},
			{command, "", "", true, 0},
			{command, "", "", false, 0},
		}...)
	}

	for _, test := range fixtures {
		t.Run(test.command+"-"+test.uuid, func(t *testing.T) {
			calls := 0
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.success && test.stderr == "" {
					assert.Equal(t, http.MethodHead, r.Method)
					assert.Equal(t, "/"+test.uuid, r.URL.Path)
					calls++
					return
				}
				if test.success && test.stderr != "" {
					defer r.Body.Close()
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/"+test.uuid, r.URL.Path)

					buffer := &bytes.Buffer{}
					_, err := buffer.ReadFrom(r.Body)
					require.NoError(t, err)
					assert.Equal(t, test.stderr, buffer.String())

					calls++
					return
				}
				if !test.success {
					defer r.Body.Close()
					assert.Equal(t, http.MethodPost, r.Method)
					assert.Equal(t, "/"+test.uuid+"/fail", r.URL.Path)

					buffer := &bytes.Buffer{}
					_, err := buffer.ReadFrom(r.Body)
					require.NoError(t, err)
					assert.Equal(t, "error during test\n\n"+test.stderr, buffer.String())

					calls++
					return
				}
			})
			server := httptest.NewServer(handler)

			healthchecks := NewProgress(&config.Profile{
				HealthChecksURL: server.URL,
				Backup: &config.BackupSection{
					HealthChecksUUID: test.uuid,
				},
			})

			var err error
			if !test.success {
				err = errors.New("error during test")
			}
			healthchecks.Summary(test.command, monitor.Summary{}, test.stderr, err)
			assert.Equal(t, test.expectedCalls, calls)
		})
	}
}
