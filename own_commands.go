package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// commandContext is the context for running a command.
type commandContext struct {
	Context
	ownCommands *OwnCommands
}

type ownCommand struct {
	name              string
	description       string
	longDescription   string
	pre               func(*Context) error                  // pre-command action (for checking the context)
	action            func(io.Writer, commandContext) error // run command action
	needConfiguration bool                                  // true if the action needs a configuration file loaded
	hide              bool                                  // don't display the command in help and completion
	hideInCompletion  bool                                  // don't display the command in completion
	noProfile         bool                                  // true if the command doesn't need a profile name
	flags             map[string]string                     // own command flags should be simple enough to be handled manually for now
}

// OwnCommands is a list of resticprofile commands
type OwnCommands struct {
	commands []ownCommand
}

func NewOwnCommands() *OwnCommands {
	return &OwnCommands{
		commands: make([]ownCommand, 0, 20),
	}
}

func (o *OwnCommands) Register(commands []ownCommand) {
	o.commands = append(o.commands, commands...)
}

func (o *OwnCommands) Exists(command string, configurationLoaded bool) bool {
	for _, commandDef := range o.commands {
		if commandDef.name == command && commandDef.needConfiguration == configurationLoaded {
			return true
		}
	}
	return false
}

func (o *OwnCommands) All() []ownCommand {
	ownCommands := make([]ownCommand, len(o.commands))
	copy(ownCommands, o.commands)
	return ownCommands
}

func (o *OwnCommands) Run(ctx *Context) error {
	command := o.find(ctx.request.command)
	if command == nil {
		return fmt.Errorf("command not found: %v", ctx.request.command)
	}
	return command.action(os.Stdout, commandContext{
		ownCommands: o,
		Context:     *ctx,
	})
}

func (o *OwnCommands) Pre(ctx *Context) error {
	command := o.find(ctx.request.command)
	if command == nil {
		return fmt.Errorf("command not found: %v", ctx.request.command)
	}
	if command.pre == nil {
		return nil
	}
	return command.pre(ctx)
}

func (o *OwnCommands) find(commandName string) *ownCommand {
	commandName = strings.ToLower(commandName)
	for _, commandDef := range o.commands {
		if commandDef.name == commandName {
			return &commandDef
		}
	}
	return nil
}
