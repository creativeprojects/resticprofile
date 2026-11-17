package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobMatchesArePassedVerbatim(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"plain", "with space", "{--dry-run,x}", "$HOME", "x&y"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}
	pattern := filepath.ToSlash(filepath.Join(dir, "*"))
	written := filepath.ToSlash(filepath.Join(dir, "written"))

	matches, err := filepath.Glob(pattern)
	require.NoError(t, err)
	require.Len(t, matches, 5)

	config := `
        version = 2

        [profiles.profile]
        [profiles.profile.backup]
        source = ["` + pattern + `", "` + written + `"]

        [profiles.profile.check]
        path = true
        `
	profile, err := getResolvedProfile("toml", config, "profile")
	require.NoError(t, err)
	require.NotNil(t, profile)

	type typedValue struct {
		value   string
		argType shell.ArgType
	}
	typed := func(args []shell.Arg) (values []typedValue) {
		for _, arg := range args {
			values = append(values, typedValue{arg.Value(), arg.Type()})
		}
		return
	}
	expected := func(writtenType shell.ArgType) (values []typedValue) {
		for _, match := range matches {
			values = append(values, typedValue{filepath.FromSlash(match), shell.ArgLiteral})
		}
		return append(values, typedValue{filepath.FromSlash(written), writtenType})
	}

	t.Run("source", func(t *testing.T) {
		assert.Equal(t, expected(shell.ArgConfigBackupSource), typed(profile.GetBackupSourceArgs()))
	})

	t.Run("path", func(t *testing.T) {
		paths, found := profile.GetCommandFlags(constants.CommandCheck).Get(constants.ParameterPath)
		require.True(t, found)
		assert.Equal(t, expected(shell.ArgConfigEscape), typed(paths))
	})
}
