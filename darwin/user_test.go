//go:build darwin

package darwin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserHasUidGid(t *testing.T) {
	user := CurrentUser()
	// it is very unlikely anyone would run the tests using sudo :D
	assert.False(t, user.SudoRoot)
	assert.Greater(t, user.Uid, 500)
	assert.Greater(t, user.Gid, 0)
}
