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
	// it is very unlikely anyone would run the tests using sudo :D
	assert.False(t, user.SudoRoot)

	assert.NotEmpty(t, user.Username)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)
	assert.Equal(t, homeDir, user.HomeDir)

	if !platform.IsWindows() {
		assert.Greater(t, user.Uid, 500)
		assert.Greater(t, user.Gid, 0)
	}
	t.Logf("%+v", user)
}
