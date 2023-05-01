package main

import (
	"bytes"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

var ansiColor = func() (c *color.Color) {
	c = color.New(color.FgCyan)
	c.EnableColor()
	return
}()

var colored = ansiColor.SprintFunc()

func TestDisplayWriter(t *testing.T) {
	buffer := bytes.Buffer{}
	write := func(v ...any) string {
		buffer.Reset()
		out, closer := displayWriter(&buffer)
		out(v...)
		closer()
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
