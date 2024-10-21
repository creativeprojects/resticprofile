package config

import (
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
)

func TestGroupSchedulableCommands(t *testing.T) {
	// Define expected schedulable commands
	expectedCommands := []string{"backup", "check", "forget", "prune", "copy"}

	config := &Config{}
	group := NewGroup(config, "test")

	commands := group.SchedulableCommands()

	assert.ElementsMatch(t, commands, expectedCommands, "Schedulable commands should match expected list")

	// Test Kind() method
	assert.Equal(t, constants.SchedulableKindGroup, group.Kind(), "Group kind should be SchedulableKindGroup")
}
