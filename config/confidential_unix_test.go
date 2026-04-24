//go:build !windows

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNonConfidentialArgsFromEnvironmentVariable(t *testing.T) {
	t.Parallel()

	repo := "local:user:%s@host/path with space"
	environment := []string{
		"MY_REPOSITORY=" + fmt.Sprintf(repo, "password"),
	}
	testConfig := `
      global:
        prevent-auto-repository-file: true
      profile:
        repository: $MY_REPOSITORY
      `
	profile, err := getProfile("yaml", testConfig, "profile", "")
	assert.NoError(t, err)
	assert.NotNil(t, profile)

	args := profile.GetCommandFlags(constants.CommandBackup)
	args = args.Modify(shell.NewExpandEnvModifier(environment))
	result := GetNonConfidentialArgs(profile, args)

	expectedSecret := shell.NewArg(fmt.Sprintf(repo, "password"), shell.ArgConfigEscape).String()
	expectedPublic := shell.NewArg(fmt.Sprintf(repo, ConfidentialReplacement), shell.ArgConfigEscape).String()

	assert.Equal(t, []string{"--repo=" + expectedSecret}, args.GetAll())
	assert.Equal(t, []string{"--repo=" + expectedPublic}, result.GetAll())
}

func TestGetAutoRepositoryFileDisabledWithEnv(t *testing.T) {
	tests := []string{
		"RESTIC_REPOSITORY",
		"RESTIC_REPOSITORY_FILE",
	}

	for _, envKey := range tests {
		t.Run(envKey, func(t *testing.T) {
			config := fmt.Sprintf(`
				[my-profile]
				repository = %q
            `, defaultHttpUrl)

			defer os.Unsetenv(envKey)
			profile, err := getResolvedProfile("toml", config, "my-profile")
			require.NoError(t, err)

			hasRepoFlag := func() (found bool) {
				_, found = profile.GetCommandFlags(constants.CommandBackup).Get("repo")
				return
			}

			t.Run("no-env", func(t *testing.T) {
				require.NoError(t, os.Unsetenv(envKey))
				profile.Environment = nil
				assert.False(t, hasRepoFlag())
			})

			t.Run("profile-env", func(t *testing.T) {
				require.NoError(t, os.Unsetenv(envKey))
				profile.Environment = map[string]ConfidentialValue{
					envKey: NewConfidentialValue("1"),
				}
				assert.True(t, hasRepoFlag())
			})

			t.Run("os-env", func(t *testing.T) {
				require.NoError(t, os.Setenv(envKey, "1"))
				profile.Environment = nil
				assert.True(t, hasRepoFlag())
			})
		})
	}
}
