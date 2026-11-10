package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckAfterLoginPermission(t *testing.T) {
	t.Run("no after-login is always allowed", func(t *testing.T) {
		for _, p := range []Permission{PermissionAuto, PermissionSystem, PermissionUserBackground, PermissionUserLoggedOn} {
			assert.NoError(t, checkAfterLoginPermission(&Config{AfterLogin: false}, p))
		}
	})

	t.Run("after-login requires user_logged_on", func(t *testing.T) {
		require.NoError(t, checkAfterLoginPermission(&Config{AfterLogin: true}, PermissionUserLoggedOn))

		for _, p := range []Permission{PermissionAuto, PermissionSystem, PermissionUserBackground} {
			assert.Error(t, checkAfterLoginPermission(&Config{AfterLogin: true}, p), "permission %s should be rejected", p)
		}
	})
}
