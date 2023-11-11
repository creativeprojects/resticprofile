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

// WithConfig sets the configuration and global values. A new copy of the context is returned.
func (c *Context) WithConfig(cfg *config.Config, global *config.Global) *Context {
	newContext := c.clone()
	newContext.config = cfg
	newContext.global = global
	return newContext
}

// WithBinary sets the restic binary to use. A new copy of the context is returned.
func (c *Context) WithBinary(resticBinary string) *Context {
	newContext := c.clone()
	newContext.binary = resticBinary
	return newContext
}

// WithCommand sets the restic command. A new copy of the context is returned.
func (c *Context) WithCommand(resticCommand string) *Context {
	newContext := c.clone()
	newContext.command = resticCommand
	return newContext
}

// WithGroup sets the configuration group. A new copy of the context is returned.
func (c *Context) WithGroup(group string) *Context {
	newContext := c.clone()
	newContext.request.group = group
	return newContext
}

// WithProfile sets the profile name. A new copy of the context is returned.
// Profile and schedule information are not copied over.
func (c *Context) WithProfile(profileName string) *Context {
	newContext := c.clone()
	newContext.request.profile = profileName
	newContext.request.group = ""
	newContext.request.schedule = ""
	newContext.profile = nil
	newContext.schedule = nil
	return newContext
}

func (c *Context) clone() *Context {
	clone := *c
	return &clone
}
