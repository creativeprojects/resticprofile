//go:build darwin

package darwin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSessionType(t *testing.T) {
	fixtures := []struct {
		permission string
		expected   string
	}{
		{"", "user_logged_on"},
		{"invalid", "user_logged_on"},
		{"user_logged_on", "user_logged_on"},
		{"user", "user"},
		{"system", "system"},
	}

	for _, fixture := range fixtures {
		assert.Equal(t, fixture.expected, NewSessionType(fixture.permission).Permission())
	}
}
