package main

import (
	"os"
	"time"

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
	request       Request
	flags         commandLineFlags
	global        *config.Global
	config        *config.Config
	binary        string // where to find the restic binary
	command       string // which restic command to use
	profile       *config.Profile
	schedule      *config.Schedule // when profile is running with run-schedule command
	sigChan       chan os.Signal   // termination request
	logTarget     string           // where to send the log output
	commandOutput string           // where to send the command output when a lotTarget is set
	stopOnBattery int              // stop if running on battery
	noLock        bool             // skip profile lock file
	lockWait      time.Duration    // wait up to duration to acquire a lock
}

func CreateContext(flags commandLineFlags, global *config.Global, cfg *config.Config, ownCommands *OwnCommands) (*Context, error) {
	// The remaining arguments are going to be sent to the restic command line
	command := global.DefaultCommand
	resticArguments := flags.resticArgs
	if len(resticArguments) > 0 {
		command = resticArguments[0]
		resticArguments = resticArguments[1:]
	}

	ctx := &Context{
		request: Request{
			command:   command,
			arguments: resticArguments,
			profile:   flags.name,
			group:     "",
			schedule:  "",
		},
		flags:         flags,
		global:        global,
		config:        cfg,
		binary:        "",
		command:       "",
		profile:       nil,
		schedule:      nil,
		sigChan:       nil,
		logTarget:     global.Log, // default to global (which can be empty)
		commandOutput: global.CommandOutput,
	}
	// own commands can check the context before running
	if ownCommands.Exists(command, true) {
		err := ownCommands.Pre(ctx)
		if err != nil {
			return ctx, err
		}
	}
	// command line flag supersedes any configuration
	if flags.log != "" {
		ctx.logTarget = flags.log
	}
	if flags.commandOutput != "" {
		ctx.commandOutput = flags.commandOutput
	}
	// same for battery configuration
	if flags.ignoreOnBattery > 0 {
		ctx.stopOnBattery = flags.ignoreOnBattery
	}
	// also lock configuration
	if flags.noLock {
		ctx.noLock = true
	}
	if flags.lockWait > 0 {
		ctx.lockWait = flags.lockWait
	}
	return ctx, nil
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
