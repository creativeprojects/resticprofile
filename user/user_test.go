package user

import (
	"os"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentUser(t *testing.T) {
	user := Current()
	assert.Equal(t, os.Geteuid() == 0, user.IsRoot())

	assert.NotEmpty(t, user.Username)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, homeDir, user.UserHomeDir)

	if !platform.IsWindows() {
		assert.Greater(t, user.Uid, 500)
		assert.Greater(t, user.Gid, 0)
	}
	t.Logf("%+v", user)
}

func TestIsRoot(t *testing.T) {
	testCases := []struct {
		name     string
		user     User
		expected bool
	}{
		{
			name:     "sudo user returns true",
			user:     User{Uid: 1000, Sudo: true},
			expected: true,
		},
		{
			name:     "root user returns true",
			user:     User{Uid: 0, Sudo: false},
			expected: true,
		},
		{
			name:     "non-root and non-sudo user returns false",
			user:     User{Uid: 1000, Sudo: false},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.user.IsRoot())
		})
	}
}
