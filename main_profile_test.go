package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartProfileOrGroup(t *testing.T) {
	// Sample configuration
	configContent := `version = "2"
        [profiles.default]
         repository = "test-repo"
        [profiles.profile1]
         inherit = "default"
        [profiles.profile2]
         inherit = "default"
        [groups.group_undefined]
         profiles = ["profile1", "profile2"]
        [groups.group_true]
         profiles = ["profile1", "profile2"]
         continue-on-error = true
        [groups.group_false]
         profiles = ["profile1", "profile2"]
         continue-on-error = false
    `

	// Load configuration
	cfg, err := config.Load(bytes.NewBufferString(configContent), config.FormatTOML)
	require.NoError(t, err)

	// Mock context
	ctx := &Context{
		config: cfg,
		global: &config.Global{},
		request: Request{
			profile: "profile1",
		},
	}

	t.Run("ProfileNotFound", func(t *testing.T) {
		ctx.request.profile = "unknown"
		err := startProfileOrGroup(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrProfileNotFound)
	})

	t.Run("ProfileExists", func(t *testing.T) {
		ctx.request.profile = "profile1"
		err := startProfileOrGroup(ctx, func(ctx *Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("ProfileGroupExists", func(t *testing.T) {
		ctx.request.profile = "group_undefined"
		err := startProfileOrGroup(ctx, func(ctx *Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("ProfileGroupGlobalContinueOnErrorTrue", func(t *testing.T) {
		calls := 0
		ctx.request.profile = "group_undefined"
		ctx.global.GroupContinueOnError = true
		err := startProfileOrGroup(ctx, func(ctx *Context) error {
			calls++
			return errors.New("error")
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("ProfileGroupContinueOnErrorTrue", func(t *testing.T) {
		calls := 0
		ctx.request.profile = "group_true"
		err := startProfileOrGroup(ctx, func(ctx *Context) error {
			calls++
			return errors.New("error")
		})
		assert.NoError(t, err)
		assert.Equal(t, 2, calls)
	})

	t.Run("ProfileGroupContinueOnErrorFalse", func(t *testing.T) {
		calls := 0
		ctx.request.profile = "group_false"
		err := startProfileOrGroup(ctx, func(ctx *Context) error {
			calls++
			return errors.New("error")
		})
		assert.Error(t, err)
		assert.Equal(t, 1, calls)
	})
}
