package main

import (
	"os"

	"github.com/creativeprojects/resticprofile/config"
)

type Request struct {
	command   string   // from the command line
	arguments []string // added arguments after the restic command; all arguments for own commands
	profile   string   // profile name (if any)
	group     string   // when running as part of a group of profiles
	schedule  string   // when started with command: run-schedule <schedule-name>
}

// Context for running a profile command.
// Not everything is always available,
// but any information should be added to the context as soon as known.
type Context struct {
	request   Request
	flags     commandLineFlags
	global    *config.Global
	config    *config.Config
	binary    string // where to find the restic binary
	command   string // which restic command to use
	profile   *config.Profile
	schedule  *config.Schedule // when profile is running with run-schedule command
	sigChan   chan os.Signal   // termination request
	logTarget string           // where to send the log output
}

// WithBinary sets the restic binary to use. It doesn't create a new context.
func (c *Context) WithBinary(resticBinary string) *Context {
	c.binary = resticBinary
	return c
}

// WithCommand sets the restic command. It doesn't create a new context.
func (c *Context) WithCommand(resticCommand string) *Context {
	c.command = resticCommand
	return c
}

// WithGroup sets the configuration group. It doesn't create a new context.
func (c *Context) WithGroup(group string) *Context {
	c.request.group = group
	return c
}

// WithProfile sets the profile name. A new copy of the context is created.
// Profile and schedule information are not copied over.
func (c *Context) WithProfile(profileName string) *Context {
	return &Context{
		request: Request{
			command:   c.request.command,
			arguments: c.request.arguments,
			profile:   profileName,
			group:     "",
			schedule:  "",
		},
		flags:     c.flags,
		global:    c.global,
		config:    c.config,
		binary:    c.binary,
		command:   c.command,
		profile:   nil,
		schedule:  nil,
		sigChan:   c.sigChan,
		logTarget: c.logTarget, // the logTarget might change in case of a scheduled context :-/
	}
}
