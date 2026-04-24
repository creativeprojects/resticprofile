package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"
)

// Journey of a message sent to the terminal:
// 1. entrypoint is inputStdout/inputStderr
// 2. data is sent to colorableStdout/colorableStderr
// 3. a copy of the data is also sent to copyStdout/Stderr if enabled (final writer for the copy)
// 4. the data is finally written to the stdin/stdout writers (coming from colorable writers)

// Terminal gives access to the standard terminal input/output
type Terminal struct {
	stdin           io.Reader
	stdout          io.Writer // final writer
	stderr          io.Writer // final writer
	enableColors    maybe.Bool
	colorableStdout io.Writer // colorable writer
	colorableStderr io.Writer // colorable writer
	copyStdout      io.Writer // stdout duplicate
	copyStderr      io.Writer // stderr duplicate
	inputStdout     io.Writer // entry for stdout
	inputStderr     io.Writer // entry for stderr
}

func NewTerminal(options ...TerminalOption) *Terminal {
	t := &Terminal{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	for _, option := range options {
		option(t)
	}

	t.colorableStdout = t.getColorableWriter(t.stdout)
	t.colorableStderr = t.getColorableWriter(t.stderr)
	if t.copyStdout == nil {
		// no copy, we send to colorable directly
		t.inputStdout = t.colorableStdout
	} else {
		// send to both the output and the copy
		t.inputStdout = io.MultiWriter(t.colorableStdout, t.getColorableWriter(t.copyStdout))
	}
	if t.copyStderr == nil {
		// no copy, we send to colorable directly
		t.inputStderr = t.colorableStderr
	} else {
		// send to both the output and the copy
		t.inputStderr = io.MultiWriter(t.colorableStderr, t.getColorableWriter(t.copyStderr))
	}

	return t
}

// AskYesNo prompts the user for a message asking for a yes/no answer
func (t *Terminal) AskYesNo(message string, defaultAnswer bool) bool {
	if !strings.HasSuffix(message, "?") {
		message += "?"
	}
	var question, input string
	if defaultAnswer {
		question = "(Y/n)"
		input = "y"
	} else {
		question = "(y/N)"
		input = "n"
	}
	_, _ = t.Printf("%s %s: ", message, question)
	scanner := bufio.NewScanner(t.stdin)
	if scanner.Scan() {
		input = strings.TrimSpace(strings.ToLower(scanner.Text()))
		if len(input) > 1 {
			// take only the first character
			input = input[:1]
		}
	}

	if input == "" {
		return defaultAnswer
	}
	if input == "y" {
		return true
	}
	return false
}

// ReadPassword reads a password without echoing it to the terminal.
func (t *Terminal) ReadPassword() (string, error) {
	stdin, ok := t.stdin.(*os.File)
	if !ok || !isTerminal(stdin) {
		return t.readLine()
	}
	line, err := term.ReadPassword(fdToInt(stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(line), nil
}

// ReadLine reads some input
func (t *Terminal) readLine() (string, error) {
	buf := bufio.NewReader(t.stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// StdoutIsTerminal returns true as stdout is a terminal session
func (t *Terminal) StdoutIsTerminal() bool {
	return isTerminalWriter(t.stdout)
}

// StderrIsTerminal returns true as stderr is a terminal session
func (t *Terminal) StderrIsTerminal() bool {
	return isTerminalWriter(t.stderr)
}

func (t *Terminal) getColorableWriter(w io.Writer) io.Writer {
	if file, ok := w.(*os.File); ok && t.enableColors.IsTrueOrUndefined() && (isTerminal(file) || t.enableColors.IsTrue()) {
		return colorable.NewColorable(file)
	}
	return colorable.NewNonColorable(w)
}

// FlushAllOutput flushes all buffered output (if supported by the underlying Writer).
func (t *Terminal) FlushAllOutput() {
	for _, writer := range []io.Writer{
		t.inputStdout, t.inputStderr,
		t.copyStdout, t.copyStderr,
		t.colorableStdout, t.colorableStderr,
		t.stdout, t.stderr,
	} {
		_, _ = util.FlushWriter(writer)
	}
}

// Print formats using the default formats for its operands and writes to standard output.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func (t *Terminal) Print(a ...any) (n int, err error) {
	return fmt.Fprint(t.inputStdout, a...)
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (t *Terminal) Println(a ...any) (n int, err error) {
	return fmt.Fprintln(t.inputStdout, a...)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func (t *Terminal) Printf(format string, a ...any) (n int, err error) {
	return fmt.Fprintf(t.inputStdout, format, a...)
}

func (t *Terminal) Scanln(a ...any) (n int, err error) {
	return fmt.Fscanln(t.stdin, a...)
}

// Write implements the io.Writer interface, writing to the terminal's stdout.
func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.inputStdout.Write(p)
}

func (t *Terminal) Stdout() io.Writer {
	return t.inputStdout
}

func (t *Terminal) Stderr() io.Writer {
	return t.inputStderr
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isTerminal(file)
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	fd := fdToInt(file.Fd())
	return term.IsTerminal(fd)
}

func fdToInt(fd uintptr) int {
	return int(fd) //nolint:gosec
}

var _ io.Writer = (*Terminal)(nil)
