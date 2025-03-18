package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserHasUidGid(t *testing.T) {
	user := Current()
	// it is very unlikely anyone would run the tests using sudo :D
	assert.False(t, user.SudoRoot)
	assert.Greater(t, user.Uid, 500)
	assert.Greater(t, user.Gid, 0)
	t.Logf("%+v", user)
}
