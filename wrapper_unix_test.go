//go:build !windows

package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilteredArgumentsRegression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		format, config string
		expected       map[string][]string
	}{
		{
			format: "toml",
			config: `
				version = "1"
				
				[default]
				password-command = 'echo password'
				initialize = true
				no-error-on-warning = true
				repository = 'backup'
				
				[default.backup]
				source = [
					'test-folder',
					'test-folder-2'
				]`,
			expected: map[string][]string{
				"backup": {"backup", "--password-command=echo\\ password", "--repo=backup", "test-folder", "test-folder-2"},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			cfg, err := config.Load(strings.NewReader(test.config), test.format)
			require.NoError(t, err)
			profile, err := cfg.GetProfile("default")
			require.NoError(t, err)
			wrapper := newResticWrapper(&Context{
				flags:    commandLineFlags{dryRun: true},
				binary:   "restic",
				profile:  profile,
				command:  "test",
				terminal: term.NewTerminal(),
			})

			for command, commandline := range test.expected {
				args := profile.GetCommandFlags(command)
				cmd := wrapper.prepareCommand(command, args, true)

				assert.Equal(t, commandline, cmd.args)
			}
		})
	}
}

func TestPrepareCommandShouldEscapeBinary(t *testing.T) {
	t.Parallel()

	profile := config.NewProfile(&config.Config{}, "name")
	ctx := &Context{
		binary:   "/full path to/restic",
		profile:  profile,
		command:  "backup",
		terminal: term.NewTerminal(),
	}
	wrapper := newResticWrapper(ctx)
	args := shell.NewArgs()
	cmd := wrapper.prepareCommand("backup", args, false)
	assert.Equal(t, `/full\ path\ to/restic`, cmd.command)
}
