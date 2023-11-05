package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestContextWithBinary(t *testing.T) {
	ctx := &Context{
		binary: "test",
	}
	ctx = ctx.WithBinary("test2")
	assert.Equal(t, "test2", ctx.binary)
}

func TestContextWithCommand(t *testing.T) {
	ctx := &Context{
		command: "test",
	}
	ctx = ctx.WithCommand("test2")
	assert.Equal(t, "test2", ctx.command)
}

func TestContextWithGroup(t *testing.T) {
	ctx := &Context{
		request: Request{
			command: "test",
			group:   "test",
		},
	}
	ctx = ctx.WithGroup("test2")
	assert.Equal(t, "test2", ctx.request.group)
	assert.NotEmpty(t, ctx.request.command)
}

func TestContextWithProfile(t *testing.T) {
	ctx := &Context{
		request: Request{
			command:  "test",
			profile:  "test",
			group:    "test",
			schedule: "test",
		},
		profile:  &config.Profile{},
		schedule: &config.Schedule{},
	}
	ctx = ctx.WithProfile("test2")
	assert.Equal(t, "test2", ctx.request.profile)
	assert.NotEmpty(t, ctx.request.command)

	assert.Empty(t, ctx.request.group)
	assert.Empty(t, ctx.request.schedule)

	assert.Nil(t, ctx.profile)
	assert.Nil(t, ctx.schedule)
}
