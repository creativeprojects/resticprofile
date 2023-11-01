package main

import (
	"os"

	"github.com/creativeprojects/resticprofile/config"
)

// Context for running a profile command.
// Not everything is always available,
// but any information should be added to the context as soon as known.
type Context struct {
	flags         commandLineFlags
	global        *config.Global
	config        *config.Config
	resticBinary  string
	resticCommand string
	arguments     []string // arguments added after the restic command || all arguments for own commands
	profileName   string
	group         string // when running as part of a group of profiles
	scheduleName  string // when started with command: run-schedule <schedule-name>
	profile       *config.Profile
	schedule      *config.Schedule // when profile is running with run-schedule command
	sigChan       chan os.Signal
}

// NewProfileContext creates a new context from the current one.
// It does NOT copy the profile and schedule.
func (c *Context) NewProfileContext() *Context {
	return &Context{
		flags:         c.flags,
		global:        c.global,
		config:        c.config,
		resticBinary:  c.resticBinary,
		resticCommand: c.resticCommand,
		arguments:     c.arguments,
		profileName:   c.profileName,
		group:         c.group,
		scheduleName:  c.scheduleName,
		profile:       nil,
		schedule:      nil,
		sigChan:       c.sigChan,
	}
}
