package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"slices"
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLiterals(t *testing.T) {
	t.Parallel()

	render := func(value string) string { return "<" + value + ">" }
	tests := []struct{ arg, expected string }{
		{"plain", "plain"},
		{markLiteral("a b"), "<a b>"},
		{"--path=" + markLiteral("a"), "--path=<a>"},
		{markLiteral("a") + "," + markLiteral("b"), "<a>,<b>"},
		{markLiteral(""), "<>"},
		{"unbalanced\x00tail", "unbalanced<tail>"},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, resolveLiterals(test.arg, render))
	}
}

func TestQuotePosixLiteral(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `''`, quotePosixLiteral(""))
	assert.Equal(t, `'a b'`, quotePosixLiteral("a b"))
	assert.Equal(t, `'it'\''s'`, quotePosixLiteral("it's"))
}

func TestEscapeWindowsArg(t *testing.T) {
	t.Parallel()

	tests := []struct{ value, expected string }{
		{"", `""`},
		{"plain", "plain"},
		{`C:\share\0&`, `C:\share\0&`},
		{"with space", `"with space"`},
		{"tab\there", "\"tab\there\""},
		{`C:\with space\`, `"C:\with space\\"`},
		{`a"b`, `"a\"b"`},
		{`a\"b`, `"a\\\"b"`},
	}
	for _, test := range tests {
		assert.Equal(t, test.expected, escapeWindowsArg(test.value), test.value)
	}
}

func TestLiteralArgumentsAreNotModified(t *testing.T) {
	t.Parallel()

	arg := NewArg("$HOME", ArgLiteral)

	expanded, changed := NewExpandEnvModifier([]string{"HOME=/home"}).Arg("path", &arg)
	assert.False(t, changed)
	assert.Equal(t, "$HOME", expanded.Value())

	legacy, changed := NewLegacyArgModifier(true).Arg("path", &arg)
	assert.False(t, changed)
	assert.Equal(t, ArgLiteral, legacy.Type())
}

func TestGetAllForCommand(t *testing.T) {
	t.Parallel()

	args := NewArgs()
	args.AddFlag("path", NewArg("/a b", ArgLiteral))
	args.AddFlag("tag", NewArg("x y", ArgConfigEscape))
	args.AddArg(NewArg("/c d", ArgLiteral))

	tag := NewArg("x y", ArgConfigEscape).String()
	assert.Equal(t, []string{"--path=\x00/a b\x00", "--tag=" + tag, "\x00/c d\x00"}, args.GetAllForCommand())
	assert.Equal(t, []string{
		"--path=" + NewArg("/a b", ArgLiteral).String(),
		"--tag=" + tag,
		NewArg("/c d", ArgLiteral).String(),
	}, args.GetAll())
}

func TestLiteralArgumentsComposing(t *testing.T) {
	t.Parallel()

	arguments := []string{"backup", "--path=" + markLiteral("/a b"), markLiteral("$HOME&x")}
	posix := []string{"-c", mockBinary + ` backup --path='/a b' '$HOME&x'`}

	tests := []struct {
		shell                     string
		expectedArgs, expectedEnv []string
	}{
		{shell: bashShell, expectedArgs: posix},
		{shell: defaultShell, expectedArgs: posix},
		{
			shell:        windowsShell,
			expectedArgs: []string{"/V:ON", "/C", mockBinary, "backup", "--path=!RESTICPROFILE_LITERAL_1!", "!RESTICPROFILE_LITERAL_2!"},
			expectedEnv:  []string{`RESTICPROFILE_LITERAL_1="/a b"`, "RESTICPROFILE_LITERAL_2=$HOME&x"},
		},
		{
			shell:        powershell,
			expectedArgs: []string{"-Command", mockBinary, `"backup"`, `"--path=${Env:RESTICPROFILE_LITERAL_1}"`, `"${Env:RESTICPROFILE_LITERAL_2}"`},
			expectedEnv:  []string{"RESTICPROFILE_LITERAL_1=/a b", "RESTICPROFILE_LITERAL_2=$HOME&x"},
		},
	}
	for _, test := range tests {
		t.Run(test.shell, func(t *testing.T) {
			t.Parallel()

			c := &Command{
				Command:   mockBinary,
				Arguments: slices.Clone(arguments),
				Environ:   []string{"HOME=/home"},
			}
			args, env := getArgumentsComposer(test.shell)(c)
			assert.Equal(t, test.expectedArgs, args)
			assert.Equal(t, test.expectedEnv, env)
		})
	}
}

// File names found on disk must reach the command exactly as they are, whatever the shell
func TestLiteralArgumentsReachCommand(t *testing.T) {
	t.Parallel()

	values := []string{
		"plain", "with space", "{--dry-run,x}", "#", "~root", "*",
		"$HOME", "${HOME}", "$(echo injected)", "`echo injected`", "a'b",
		"x;echo injected", "x&echo injected", "x|echo injected",
		"%PATH%", "!PATH!", "^caret", "(paren)", `back\slash`, `trailing\`,
	}
	shells := []string{bashShell, defaultShell}
	if platform.IsWindows() {
		shells = []string{windowsShell, powershell, powershell6}
	} else {
		values = append(values, `a"b`, "new\nline")
	}

	args := NewArgs()
	args.AddFlags("path", NewArgsSlice(values, ArgLiteral))
	args.AddArgs(NewArgsSlice(values, ArgLiteral))

	expected := make([]string, 0, len(values)*2)
	for _, value := range values {
		expected = append(expected, "--path="+value)
	}
	expected = append(expected, values...)

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			t.Parallel()
			if _, err := exec.LookPath(shell); err != nil {
				t.Skipf("shell %q not available", shell)
			}

			stdout := &bytes.Buffer{}
			c := NewCommand(echoBinary, args.GetAllForCommand())
			c.Shell = []string{shell}
			c.Stdout = stdout

			_, _, err := c.Run()
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("%q\n", expected), stdout.String())
		})
	}
}
