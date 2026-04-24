package main

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

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
