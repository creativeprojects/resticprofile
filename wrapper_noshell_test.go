package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResticWithoutShell(t *testing.T) {
	t.Setenv("TEST_NO_SHELL_TAG", "tag value")

	cfg, err := config.Load(strings.NewReader(`
version: "1"
global:
  restic-without-shell: true
default:
  repository: "local:/tmp/repo"
  backup:
    exclude: ["My Documents", "My\\ Documents", "*.tmp", "$$HOME"]
    tag: "$TEST_NO_SHELL_TAG"
    option: "$TEST_NO_SHELL_UNDEFINED"
    source: ["/source dir", "/Côte d'Ivoire", "/path; echo foo", "/{a,b}"]
`), "yaml")
	require.NoError(t, err)
	global, err := cfg.GetGlobalSection()
	require.NoError(t, err)
	require.True(t, global.ResticWithoutShell)
	profile, err := cfg.GetProfile("default")
	require.NoError(t, err)

	expected := fmt.Sprintf("%q", []string{
		"backup",
		"--exclude=My Documents", `--exclude=My\ Documents`, "--exclude=*.tmp", "--exclude=$HOME",
		"--option=$TEST_NO_SHELL_UNDEFINED",
		"--repo=local:/tmp/repo",
		"--tag=tag value",
		filepath.FromSlash("/source dir"), filepath.FromSlash("/Côte d'Ivoire"),
		filepath.FromSlash("/path; echo foo"), filepath.FromSlash("/{a,b}"),
	})

	// legacy arguments only change the escaping for the shell
	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy=%v", legacy), func(t *testing.T) {
			buffer := new(bytes.Buffer)
			ctx := &Context{
				request:    Request{profile: "default"},
				global:     global,
				binary:     echoBinary,
				profile:    profile,
				command:    "backup",
				legacyArgs: legacy,
				terminal:   term.NewTerminal(term.WithStdout(buffer)),
			}
			require.NoError(t, newResticWrapper(ctx).runCommand("backup"))
			assert.Equal(t, expected, strings.TrimSpace(buffer.String()))
		})
	}
}
