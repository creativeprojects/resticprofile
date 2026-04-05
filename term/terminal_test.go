package term

import (
	"bytes"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalAskYesNo(t *testing.T) {
	t.Parallel()

	fixtures := []struct {
		input         string
		defaultAnswer bool
		expected      bool
	}{
		// Empty answer => will follow the defaultAnswer
		{"", true, true},
		{"", false, false},
		{"\n", true, true},
		{"\n", false, false},
		{"\r\n", true, true},
		{"\r\n", false, false},
		// Garbage answer => will always return false
		{"aa", true, false},
		{"aa", false, false},
		{"aa\n", true, false},
		{"aa\n", false, false},
		{"aa\r\n", true, false},
		{"aa\r\n", false, false},
		// Answer yes
		{"y", true, true},
		{"y", false, true},
		{"y\n", true, true},
		{"y\n", false, true},
		{"y\r\n", true, true},
		{"y\r\n", false, true},
		// Full answer yes
		{"yes", true, true},
		{"yes", false, true},
		{"yes\n", true, true},
		{"yes\n", false, true},
		{"yes\r\n", true, true},
		{"yes\r\n", false, true},
		// Answer no
		{"n", true, false},
		{"n", false, false},
		{"n\n", true, false},
		{"n\n", false, false},
		{"n\r\n", true, false},
		{"n\r\n", false, false},
		// Full answer no
		{"no", true, false},
		{"no", false, false},
		{"no\n", true, false},
		{"no\n", false, false},
		{"no\r\n", true, false},
		{"no\r\n", false, false},
	}
	for _, fixture := range fixtures {
		output := new(bytes.Buffer)
		terminal := NewTerminal(WithStdin(bytes.NewBufferString(fixture.input)), WithStdout(output))
		result := terminal.AskYesNo("message", fixture.defaultAnswer)
		assert.Contains(t, output.String(), "message? (")
		assert.Equalf(t, fixture.expected, result, "when input was %q", fixture.input)
	}
}

func ExamplePrint() {
	terminal := NewTerminal()
	_, err := terminal.Print("ExampleTerminalPrint")
	if err != nil {
		log.Fatal(err)
	}
	// Output: ExampleTerminalPrint
}

func TestDefaultTerminal(t *testing.T) {
	terminal := NewTerminal()
	// in a test environment, stdout and stderr are not terminals, so we expect false for both
	assert.False(t, terminal.StdoutIsTerminal())
	assert.False(t, terminal.StderrIsTerminal())
}

func TestReadPasswordFromBuffer(t *testing.T) {
	input := "mysecretpassword\n"
	terminal := NewTerminal(WithStdin(bytes.NewBufferString(input)))
	password, err := terminal.ReadPassword()
	assert.NoError(t, err)
	assert.Equal(t, "mysecretpassword", password)
}

func TestTerminalOutputCapture(t *testing.T) {
	buffer := new(bytes.Buffer)
	recorder, err := NewRecorder(buffer)
	require.NoError(t, err)

	terminal := NewTerminal(WithStdoutRecorder(recorder))
	written, err := terminal.Print("TestTerminalOutputCapture")
	require.NoError(t, err)
	assert.Equal(t, 25, written)

	err = recorder.Close()
	require.NoError(t, err)
	assert.Equal(t, "TestTerminalOutputCapture", buffer.String())
}

// ansiText is a string with ANSI color codes that should be passed through or stripped depending on the writer.
const ansiText = "\x1b[31mhello\x1b[0m"
const plainText = "hello"

func TestGetColorableWriterWithNonFileWriterReturnsNonColorable(t *testing.T) {
	t.Parallel()
	// A bytes.Buffer is not an *os.File, so getColorableWriter always wraps it as NonColorable.
	buf := &bytes.Buffer{}
	terminal := NewTerminal()
	writer := terminal.getColorableWriter(buf)

	_, err := writer.Write([]byte(ansiText))
	require.NoError(t, err)

	// NonColorable strips ANSI escape codes.
	assert.Equal(t, plainText, buf.String())
}

func TestGetColorableWriterWithColorsDisabledReturnsNonColorable(t *testing.T) {
	t.Parallel()
	// Even for an *os.File, explicitly disabling colors forces a NonColorable writer.
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()

	terminal := NewTerminal(WithColors(false))
	writer := terminal.getColorableWriter(pw)

	_, err = writer.Write([]byte(ansiText))
	require.NoError(t, err)
	pw.Close()

	out, err := io.ReadAll(pr)
	require.NoError(t, err)
	// NonColorable strips ANSI escape codes.
	assert.Equal(t, plainText, string(out))
}

func TestGetColorableWriterWithColorsUndefinedOnNonTerminalFileReturnsNonColorable(t *testing.T) {
	t.Parallel()
	// With undefined colors and a non-terminal *os.File (e.g. a pipe), the writer is NonColorable.
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()

	terminal := NewTerminal() // enableColors is undefined
	writer := terminal.getColorableWriter(pw)

	_, err = writer.Write([]byte(ansiText))
	require.NoError(t, err)
	pw.Close()

	out, err := io.ReadAll(pr)
	require.NoError(t, err)
	// NonColorable strips ANSI escape codes.
	assert.Equal(t, plainText, string(out))
}

func TestGetColorableWriterWithColorsEnabledOnNonTerminalFileReturnsColorable(t *testing.T) {
	t.Parallel()
	// With colors explicitly enabled, getColorableWriter returns a Colorable writer even for a
	// non-terminal *os.File, so ANSI codes are passed through unchanged.
	pr, pw, err := os.Pipe()
	require.NoError(t, err)
	defer pr.Close()

	terminal := NewTerminal(WithColors(true))
	writer := terminal.getColorableWriter(pw)

	_, err = writer.Write([]byte(ansiText))
	require.NoError(t, err)
	pw.Close()

	out, err := io.ReadAll(pr)
	require.NoError(t, err)
	// Colorable preserves ANSI escape codes.
	assert.Equal(t, ansiText, string(out))
}
