package main

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ansiColor = func() (c *color.Color) {
	c = color.New(color.FgCyan)
	c.EnableColor()
	return
}()

var colored = ansiColor.SprintFunc()

func TestDisplayWriter(t *testing.T) {
	write := func(v ...any) string {
		buffer := new(bytes.Buffer)
		recorder, err := term.NewRecorder(buffer)
		require.NoError(t, err)

		terminal := term.NewTerminal(term.WithStdoutRecorder(recorder), term.WithColors(true))
		out, closer := displayWriter(terminal)
		out(v...)
		closer()
		recorder.Close()
		return buffer.String()
	}

	t.Run("write-plain", func(t *testing.T) {
		actual := write("hello %s %d")
		assert.Equal(t, "hello %s %d", actual)
	})

	t.Run("write-with-format", func(t *testing.T) {
		actual := write("hello %s %02d", "world", 5)
		assert.Equal(t, "hello world 05", actual)
	})

	t.Run("write-tabs", func(t *testing.T) {
		actual := write("col1\tcol2\tcol3\nvalue1\tvalue2\tvalue3")
		assert.Equal(t, "col1    col2    col3\nvalue1  value2  value3", actual)
	})

	t.Run("ansi", func(t *testing.T) {
		actual := write(colored("test"))
		assert.Equal(t, colored("test"), actual)
		assert.Contains(t, colored("test"), "\x1b[")
	})
}

func TestDisplayWriterNoColors(t *testing.T) {
	write := func(v ...any) string {
		buffer := new(bytes.Buffer)
		recorder, err := term.NewRecorder(buffer)
		require.NoError(t, err)

		terminal := term.NewTerminal(term.WithStdoutRecorder(recorder), term.WithColors(false))
		out, closer := displayWriter(terminal)
		out(v...)
		closer()
		recorder.Close()
		return buffer.String()
	}
	actual := write(colored("test"))
	assert.Equal(t, "test", actual)
	assert.Contains(t, colored("test"), "\x1b[")
}

func TestDisplayVersionVerbose1(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := displayVersion(commandContext{Context: Context{terminal: term.NewTerminal(term.WithStdout(buffer)), flags: commandLineFlags{verbose: true}}})
	require.NoError(t, err)
	assert.True(t, strings.Contains(buffer.String(), runtime.GOOS))
}

func TestDisplayVersionVerbose2(t *testing.T) {
	buffer := &bytes.Buffer{}
	err := displayVersion(commandContext{Context: Context{terminal: term.NewTerminal(term.WithStdout(buffer)), request: Request{arguments: []string{"-v"}}}})
	require.NoError(t, err)
	assert.True(t, strings.Contains(buffer.String(), runtime.GOOS))
}

const profilesTestConfig = `version: "2"

groups:
  full-backup:
    description: All hosts
    profiles:
      - alpha
      - beta

profiles:
  alpha:
    description: First profile
    backup:
      source: /srv/alpha
    check:
      read-data: true
  beta:
    description: Second profile
`

func newProfilesTestContext(t *testing.T, args []string) (commandContext, *bytes.Buffer) {
	t.Helper()
	cfg, err := config.Load(bytes.NewBufferString(profilesTestConfig), config.FormatYAML)
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	terminal := term.NewTerminal(term.WithStdout(buffer), term.WithColors(false))
	return commandContext{
		Context: Context{
			config:   cfg,
			terminal: terminal,
			request:  Request{arguments: args},
		},
	}, buffer
}

func TestDisplayProfilesCommand_Plain(t *testing.T) {
	ctx, buffer := newProfilesTestContext(t, []string{"profiles"})

	require.NoError(t, displayProfilesCommand(ctx))
	out := buffer.String()

	assert.Contains(t, out, "Profiles available")
	assert.Contains(t, out, "alpha")
	assert.Contains(t, out, "beta")
	assert.Contains(t, out, "First profile")
	assert.Contains(t, out, "Second profile")
	assert.Contains(t, out, "backup, check")
	assert.Contains(t, out, "Groups available")
	assert.Contains(t, out, "full-backup")
	assert.Contains(t, out, "All hosts")
}

func TestDisplayProfilesCommand_JSON(t *testing.T) {
	ctx, buffer := newProfilesTestContext(t, []string{"profiles", "--output=json"})

	require.NoError(t, displayProfilesCommand(ctx))

	var got profilesOutput
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &got))

	require.Len(t, got.Profiles, 2)
	assert.Equal(t, profileEntry{
		Name:        "alpha",
		Description: "First profile",
		Sections:    []string{"backup", "check"},
	}, got.Profiles[0])
	assert.Equal(t, profileEntry{
		Name:        "beta",
		Description: "Second profile",
		Sections:    []string{},
	}, got.Profiles[1])
	require.Len(t, got.Groups, 1)
	assert.Equal(t, groupEntry{
		Name:        "full-backup",
		Description: "All hosts",
		Profiles:    []string{"alpha", "beta"},
	}, got.Groups[0])
}

func TestDisplayProfilesCommand_JSONSpaceSeparated(t *testing.T) {
	ctx, buffer := newProfilesTestContext(t, []string{"profiles", "--output", "json"})

	require.NoError(t, displayProfilesCommand(ctx))

	var got profilesOutput
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &got))
	assert.NotEmpty(t, got.Profiles)
}

func TestDisplayProfilesCommand_JSONEmptyConfig(t *testing.T) {
	cfg, err := config.Load(bytes.NewBufferString("version: \"2\"\n"), config.FormatYAML)
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	terminal := term.NewTerminal(term.WithStdout(buffer), term.WithColors(false))
	ctx := commandContext{
		Context: Context{
			config:   cfg,
			terminal: terminal,
			request:  Request{arguments: []string{"profiles", "--output=json"}},
		},
	}

	require.NoError(t, displayProfilesCommand(ctx))

	// Empty slices must serialise as [] rather than null so scripts can
	// rely on the keys always being arrays.
	assert.Contains(t, buffer.String(), `"profiles": []`)
	assert.Contains(t, buffer.String(), `"groups": []`)

	var got profilesOutput
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &got))
	assert.NotNil(t, got.Profiles)
	assert.NotNil(t, got.Groups)
	assert.Empty(t, got.Profiles)
	assert.Empty(t, got.Groups)
}

func TestDisplayProfilesCommand_InvalidFormat(t *testing.T) {
	ctx, buffer := newProfilesTestContext(t, []string{"profiles", "--output=xml"})

	err := displayProfilesCommand(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "xml")
	assert.Empty(t, buffer.String())
}
