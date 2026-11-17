package shell

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoShellCommand(t *testing.T) {
	t.Parallel()

	args := []string{
		"plain", "with space", `quo"te`, "it's", "$HOME", "%PATH%", "!PATH!",
		"a&b", "a;b", "a|b", "{a,b}", "#", "*", `back\slash`, `trailing\`, "",
	}
	stdout := &bytes.Buffer{}
	c := NewCommand(echoBinary, args)
	c.NoShell = true
	c.Stdout = stdout

	command, commandArgs, err := c.GetShellCommand()
	require.NoError(t, err)
	assert.Equal(t, echoBinary, command)
	assert.Equal(t, args, commandArgs)

	_, _, err = c.Run()
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q\n", args), stdout.String())
}

func TestGetAllUnescaped(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlag("tag", NewArg("a b", ArgConfigEscape))
	args.AddFlag("exclude", NewArg("*.tmp", ArgConfigKeepGlobQuote))
	args.AddFlag("empty", NewEmptyValueArg())
	args.AddFlags("verbose", []Arg{})
	args.AddArg(NewArg("/source dir", ArgConfigBackupSource))

	assert.Equal(t, []string{"--empty=", "--exclude=*.tmp", "--tag=a b", "--verbose", "/source dir"}, args.GetAllUnescaped())
}

func TestNoShellExpandEnvModifier(t *testing.T) {
	t.Parallel()

	modifier := NewNoShellExpandEnvModifier([]string{"DEFINED=value"})
	tests := []struct {
		argType         ArgType
		value, expected string
	}{
		{ArgConfigEscape, "$DEFINED", "value"},
		{ArgConfigEscape, "${DEFINED}/x", "value/x"},
		{ArgConfigEscape, "a-$DEFINED-b", "a-value-b"},
		{ArgConfigEscape, "$UNDEFINED and ${UNDEFINED}", "$UNDEFINED and ${UNDEFINED}"},
		{ArgConfigEscape, "a$!b $$ $1 ${", "a$!b $$ $1 ${"},
		{ArgConfigKeepGlobQuote, "$DEFINED", "$DEFINED"},
		{ArgCommandLineEscape, "$DEFINED", "$DEFINED"},
		{ArgConfigBackupSource, "$DEFINED", "$DEFINED"},
	}
	for _, test := range tests {
		arg := NewArg(test.value, test.argType)
		result, changed := modifier.Arg("name", &arg)
		assert.Equal(t, test.expected, result.Value(), test.value)
		assert.Equal(t, test.expected != test.value, changed, test.value)
	}
}
