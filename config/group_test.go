package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupSchedulableCommands(t *testing.T) {
	group := NewGroup(nil, "test")
	commands := group.SchedulableCommands()
	assert.GreaterOrEqual(t, len(commands), 5)
}
