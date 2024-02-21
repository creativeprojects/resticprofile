package main

import (
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/stretchr/testify/assert"
)

func TestContextClone(t *testing.T) {
	ctx := &Context{
		config: &config.Config{},
		global: &config.Global{},
		binary: "test",
	}
	clone := ctx.clone()
	assert.False(t, ctx == clone) // different pointers
	assert.Equal(t, ctx, clone)   // same values
}

func TestContextWithConfig(t *testing.T) {
	ctx := &Context{
		config: nil,
		global: nil,
	}
	ctx = ctx.WithConfig(&config.Config{}, &config.Global{})
	assert.NotNil(t, ctx.config)
	assert.NotNil(t, ctx.global)
}

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

func TestCreateContext(t *testing.T) {
	fixtures := []struct {
		description string
		flags       commandLineFlags
		global      *config.Global
		cfg         *config.Config
		ownCommands *OwnCommands
		context     *Context
	}{
		{
			description: "empty config with default command only",
			flags:       commandLineFlags{},
			global: &config.Global{
				DefaultCommand: "test",
			},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{},
			context: &Context{
				request: Request{
					command: "test",
				},
				global: &config.Global{
					DefaultCommand: "test",
				},
				config: &config.Config{},
			},
		},
		{
			description: "command and arguments",
			flags: commandLineFlags{
				resticArgs: []string{"arg1", "arg2"},
			},
			global: &config.Global{
				DefaultCommand: "test",
			},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{},
			context: &Context{
				flags: commandLineFlags{
					resticArgs: []string{"arg1", "arg2"},
				},
				request: Request{
					command:   "arg1",
					arguments: []string{"arg2"},
				},
				global: &config.Global{
					DefaultCommand: "test",
				},
				config: &config.Config{},
			},
		},
		{
			description: "global log target",
			flags:       commandLineFlags{},
			global:      &config.Global{Log: "global"},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{},
			context: &Context{
				flags:     commandLineFlags{},
				request:   Request{},
				global:    &config.Global{Log: "global"},
				config:    &config.Config{},
				logTarget: "global",
			},
		},
		{
			description: "log target on the command line",
			flags:       commandLineFlags{log: "cmdline"},
			global:      &config.Global{Log: "global"},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{},
			context: &Context{
				flags:     commandLineFlags{log: "cmdline"},
				request:   Request{},
				global:    &config.Global{Log: "global"},
				config:    &config.Config{},
				logTarget: "cmdline",
			},
		},
		{
			description: "log target from command",
			flags:       commandLineFlags{},
			global:      &config.Global{Log: "global", DefaultCommand: "test"},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{
				commands: []ownCommand{
					{
						name:              "test",
						needConfiguration: true,
						pre: func(ctx *Context) error {
							ctx.logTarget = "command"
							return nil
						},
					},
				},
			},
			context: &Context{
				flags:     commandLineFlags{},
				request:   Request{command: "test"},
				global:    &config.Global{Log: "global", DefaultCommand: "test"},
				config:    &config.Config{},
				logTarget: "command",
			},
		},
		{
			description: "log target from command line",
			flags: commandLineFlags{
				log:        "cmdline",
				resticArgs: []string{"test"},
			},
			global: &config.Global{},
			cfg:    &config.Config{},
			ownCommands: &OwnCommands{
				commands: []ownCommand{
					{
						name:              "test",
						needConfiguration: true,
						pre: func(ctx *Context) error {
							ctx.logTarget = "command"
							return nil
						},
					},
				},
			},
			context: &Context{
				flags: commandLineFlags{
					log:        "cmdline",
					resticArgs: []string{"test"},
				},
				request: Request{
					command:   "test",
					arguments: []string{},
				},
				global:    &config.Global{},
				config:    &config.Config{},
				logTarget: "cmdline",
			},
		},
		{
			description: "battery and lock from command",
			flags:       commandLineFlags{},
			global:      &config.Global{DefaultCommand: "test"},
			cfg:         &config.Config{},
			ownCommands: &OwnCommands{
				commands: []ownCommand{
					{
						name:              "test",
						needConfiguration: true,
						pre: func(ctx *Context) error {
							ctx.stopOnBattery = 10
							ctx.lockWait = 20
							ctx.noLock = true
							return nil
						},
					},
				},
			},
			context: &Context{
				flags:         commandLineFlags{},
				request:       Request{command: "test"},
				global:        &config.Global{DefaultCommand: "test"},
				config:        &config.Config{},
				stopOnBattery: 10,
				lockWait:      20,
				noLock:        true,
			},
		},
		{
			description: "battery and lock from command line",
			flags: commandLineFlags{
				resticArgs:      []string{"test"},
				ignoreOnBattery: 80,
				lockWait:        30,
				noLock:          true,
			},
			global: &config.Global{},
			cfg:    &config.Config{},
			ownCommands: &OwnCommands{
				commands: []ownCommand{
					{
						name:              "test",
						needConfiguration: true,
						pre: func(ctx *Context) error {
							ctx.stopOnBattery = 10
							ctx.lockWait = 20
							return nil
						},
					},
				},
			},
			context: &Context{
				flags: commandLineFlags{
					resticArgs:      []string{"test"},
					ignoreOnBattery: 80,
					lockWait:        30,
					noLock:          true,
				},
				request: Request{
					command:   "test",
					arguments: []string{},
				},
				global:        &config.Global{},
				config:        &config.Config{},
				stopOnBattery: 80,
				lockWait:      30,
				noLock:        true,
			},
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.description, func(t *testing.T) {
			ctx, err := CreateContext(fixture.flags, fixture.global, fixture.cfg, fixture.ownCommands)
			assert.NoError(t, err)
			assert.Equal(t, fixture.context, ctx)
		})
	}
}
