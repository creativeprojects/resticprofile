package user

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestCurrentUser(t *testing.T) {
	user := Current()
	// it is very unlikely anyone would run the tests using sudo :D
	assert.False(t, user.SudoRoot)

	assert.NotEmpty(t, user.Username)

	if !platform.IsWindows() {
		assert.Greater(t, user.Uid, 500)
		assert.Greater(t, user.Gid, 0)
	}
	t.Logf("%+v", user)
}
