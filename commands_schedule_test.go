package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const scheduleIntegrationTestsConfiguration = `
version: "2"

global:
  scheduler: crontab:*:%s

groups:
  two-profiles:
    profiles:
      - profile-schedule-inline
      - profile-schedule-struct

profiles:
  default-inline:
    backup:
      schedule-permission: "user"

  default-struct:
    backup:
      schedule:
        permission: "user"

  profile-schedule-inline:
    backup:
      schedule: daily

  profile-schedule-struct:
    backup:
      schedule:
        at: daily

`

func TestCommandsIntegrationUsingCrontab(t *testing.T) {
	crontab := filepath.Join(t.TempDir(), "crontab")
	cfg, err := config.Load(
		bytes.NewBufferString(fmt.Sprintf(scheduleIntegrationTestsConfiguration, crontab)),
		config.FormatYAML,
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	testCases := []struct {
		name     string
		contains string
		err      error
	}{
		{
			name: "",
			err:  config.ErrNotFound,
		},
		{
			name:     "profile-schedule-inline",
			contains: "Original form: daily",
		},
		{
			name:     "profile-schedule-struct",
			contains: "Original form: daily",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := commandContext{
				Context: Context{
					config: cfg,
					flags: commandLineFlags{
						name: tc.name,
					},
				},
			}
			output := &bytes.Buffer{}
			term.SetOutput(output)
			defer term.SetOutput(os.Stdout)

			err = statusSchedule(output, ctx)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)

			assert.Contains(t, output.String(), tc.contains)
			t.Log(output.String())
		})
	}
}
