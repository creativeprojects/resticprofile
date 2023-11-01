package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestCloningContextLeavesProfileAndScheduleBehind(t *testing.T) {
	ctx := Context{
		profile:  &config.Profile{},
		schedule: &config.Schedule{},
	}
	newCtx := ctx.NewProfileContext()
	assert.Nil(t, newCtx.profile)
	assert.Nil(t, newCtx.schedule)
}
