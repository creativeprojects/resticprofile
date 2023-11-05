package main

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func fakeCommands() *OwnCommands {
	ownCommands := NewOwnCommands()
	ownCommands.Register([]ownCommand{
		{
			name:              "first",
			description:       "first first",
			pre:               pre,
			action:            firstCommand,
			needConfiguration: false,
		},
		{
			name:              "second",
			description:       "second second",
			action:            secondCommand,
			needConfiguration: true,
			flags: map[string]string{
				"-f, --first":   "first flag",
				"-s, --seccond": "second flag",
			},
		},
		{
			name:              "third",
			description:       "third third",
			action:            thirdCommand,
			needConfiguration: false,
			hide:              true,
		},
	})
	return ownCommands
}

func firstCommand(_ io.Writer, _ commandContext) error {
	return errors.New("first")
}

func secondCommand(_ io.Writer, _ commandContext) error {
	return errors.New("second")
}

func thirdCommand(_ io.Writer, _ commandContext) error {
	return errors.New("third")
}

func pre(_ *Context) error {
	return errors.New("pre")
}

func TestDisplayOwnCommands(t *testing.T) {
	buffer := &strings.Builder{}
	displayOwnCommands(buffer, commandContext{ownCommands: fakeCommands()})
	assert.Equal(t, "  first   first first\n  second  second second\n", buffer.String())
}

func TestDisplayOwnCommand(t *testing.T) {
	buffer := &strings.Builder{}
	displayOwnCommandHelp(buffer, "second", commandContext{ownCommands: fakeCommands()})
	assert.Equal(t, `Purpose: second second

Usage:
  resticprofile [resticprofile flags] [profile name.]second [command specific flags]

Flags:
  -f, --first    first flag
  -s, --seccond  second flag

`, buffer.String())
}

func TestIsOwnCommand(t *testing.T) {
	assert.True(t, fakeCommands().Exists("first", false))
	assert.True(t, fakeCommands().Exists("second", true))
	assert.True(t, fakeCommands().Exists("third", false))
	assert.False(t, fakeCommands().Exists("another one", true))
}

func TestRunOwnCommand(t *testing.T) {
	assert.EqualError(t, fakeCommands().Run(&Context{request: Request{command: "first"}}), "first")
	assert.EqualError(t, fakeCommands().Run(&Context{request: Request{command: "second"}}), "second")
	assert.EqualError(t, fakeCommands().Run(&Context{request: Request{command: "third"}}), "third")
	assert.EqualError(t, fakeCommands().Run(&Context{request: Request{command: "another one"}}), "command not found: another one")
}

func TestPreOwnCommand(t *testing.T) {
	assert.EqualError(t, fakeCommands().Pre(&Context{request: Request{command: "first"}}), "pre")
	assert.NoError(t, fakeCommands().Pre(&Context{request: Request{command: "second"}}))
	assert.NoError(t, fakeCommands().Pre(&Context{request: Request{command: "third"}}))
	assert.EqualError(t, fakeCommands().Pre(&Context{request: Request{command: "another one"}}), "command not found: another one")
}
