package main

import (
	"fmt"
	"io"
	"os"

	"github.com/creativeprojects/resticprofile/config"
)

type commandRequest struct {
	ownCommands *OwnCommands
	config      *config.Config
	flags       commandLineFlags
	args        []string
}

type ownCommand struct {
	name              string
	description       string
	longDescription   string
	action            func(io.Writer, commandRequest) error
	needConfiguration bool              // true if the action needs a configuration file loaded
	hide              bool              // don't display the command in help and completion
	hideInCompletion  bool              // don't display the command in completion
	flags             map[string]string // own command flags should be simple enough to be handled manually for now
}

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

func (o *OwnCommands) Run(configuration *config.Config, commandName string, flags commandLineFlags, args []string) error {
	for _, command := range o.commands {
		if command.name == commandName {
			return command.action(os.Stdout, commandRequest{
				ownCommands: o,
				config:      configuration,
				flags:       flags,
				args:        args,
			})
		}
	}
	return fmt.Errorf("command not found: %v", commandName)
}
