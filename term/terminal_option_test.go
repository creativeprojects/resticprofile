package term

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalNoInput(t *testing.T) {
	terminal := NewTerminal(WithNoStdin())
	pwd, err := terminal.ReadPassword()
	require.Error(t, err)
	assert.Empty(t, pwd)

	var line string
	read, err := terminal.Scanln(&line)
	require.Error(t, err)
	assert.Empty(t, line)
	assert.Empty(t, read)
}

func TestTerminalNoStdout(t *testing.T) {
	terminal := NewTerminal(WithNoStdout())
	written, err := terminal.Print("something")
	require.NoError(t, err)
	assert.Equal(t, 9, written)
}

func TestTerminalNoStderr(t *testing.T) {
	terminal := NewTerminal(WithNoStderr())
	written, err := fmt.Fprint(terminal.Stderr(), "something")
	require.NoError(t, err)
	assert.Equal(t, 9, written)
}

func TestTestTerminalStdoutCopy(t *testing.T) {
	buffer := new(bytes.Buffer)
	terminal := NewTerminal(WithStdoutCopy(buffer))
	_, err := terminal.Printf("%s test", "copy")
	require.NoError(t, err)
	assert.Equal(t, "copy test", buffer.String())
}
