package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// File names found by a glob pattern must reach restic verbatim, whatever the shell
func TestGlobMatchesReachResticVerbatim(t *testing.T) {
	// relative source globs are resolved from the current directory
	dir, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
	t.Chdir(dir)

	names := []string{
		"data", "{--dry-run,x}", "#", "$HOME", "a'b", "(paren)", "^caret",
		"x;echo injected", "x&echo injected", "x|echo injected", "%PATH%", "!PATH!",
	}
	for _, name := range names {
		require.NoError(t, os.WriteFile(name, nil, 0o600))
	}
	matches, err := filepath.Glob("*")
	require.NoError(t, err)
	require.Len(t, matches, len(names))

	cfg, err := config.Load(strings.NewReader(fmt.Sprintf(`
version: "2"
profiles:
  default:
    repository: "local:/tmp/repo"
    backup:
      source-relative: true
      source-base: %q
      source: ["*"]
    check:
      path: true
`, filepath.ToSlash(dir))), "yaml")
	require.NoError(t, err)
	profile, err := cfg.GetProfile("default")
	require.NoError(t, err)

	run := func(t *testing.T, command string) string {
		t.Helper()
		buffer := new(bytes.Buffer)
		ctx := &Context{
			request:  Request{profile: "default"},
			binary:   echoBinary,
			profile:  profile,
			command:  command,
			terminal: term.NewTerminal(term.WithStdout(buffer)),
		}
		require.NoError(t, newResticWrapper(ctx).runCommand(command))
		return strings.TrimSpace(buffer.String())
	}

	t.Run("backup", func(t *testing.T) {
		expected := append([]string{"backup", "--host", "--repo=local:/tmp/repo"}, matches...)
		assert.Equal(t, fmt.Sprintf("%q", expected), run(t, "backup"))
	})

	t.Run("path", func(t *testing.T) {
		expected := make([]string, 0, len(matches)+2)
		expected = append(expected, "check")
		for _, match := range matches {
			expected = append(expected, "--path="+filepath.Join(dir, match))
		}
		expected = append(expected, "--repo=local:/tmp/repo")
		assert.Equal(t, fmt.Sprintf("%q", expected), run(t, "check"))
	})
}
