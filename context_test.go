package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

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
