package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mattn/go-colorable"
	"golang.org/x/term"
)

var (
	terminalOutput io.Writer = os.Stdout
	errorOutput    io.Writer = os.Stderr
)

// Flusher allows a Writer to declare it may buffer content that can be flushed
type Flusher interface {
	// Flush writes any pending bytes to output
	Flush() error
}

// AskYesNo prompts the user for a message asking for a yes/no answer
func AskYesNo(reader io.Reader, message string, defaultAnswer bool) bool {
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
	fmt.Printf("%s %s: ", message, question)
	scanner := bufio.NewScanner(reader)
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
func ReadPassword() (string, error) {
	stdin := int(os.Stdin.Fd())
	if !term.IsTerminal(stdin) {
		return ReadLine()
	}
	line, err := term.ReadPassword(stdin)
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(line), nil
}

// ReadLine reads some input
func ReadLine() (string, error) {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// OsStdoutIsTerminal returns true as os.Stdout is a terminal session
func OsStdoutIsTerminal() bool {
	fd := int(os.Stdout.Fd())
	return term.IsTerminal(fd)
}

// OsStdoutIsTerminal returns true as os.Stdout is a terminal session
func OsStdoutTerminalSize() (width, height int) {
	fd := int(os.Stdout.Fd())
	var err error
	width, height, err = term.GetSize(fd)
	if err != nil {
		width, height = 0, 0
	}
	return
}

type LockedWriter struct {
	writer io.Writer
	mutex  *sync.Mutex
}

func (w *LockedWriter) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.writer.Write(p)
}

func (w *LockedWriter) Flush() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if f, ok := w.writer.(Flusher); ok {
		err = f.Flush()
	}
	return
}

// SetOutput changes the default output for the Print* functions
func SetOutput(w io.Writer) {
	terminalOutput = &LockedWriter{writer: w, mutex: new(sync.Mutex)}
}

// GetOutput returns the default output of the Print* functions
func GetOutput() io.Writer {
	return terminalOutput
}

// GetColorableOutput returns an output supporting ANSI color if output is a terminal
func GetColorableOutput() io.Writer {
	out := GetOutput()
	if out == os.Stdout && OsStdoutIsTerminal() {
		return colorable.NewColorable(os.Stdout)
	}
	return colorable.NewNonColorable(out)
}

// SetErrorOutput changes the error output for the Print* functions
func SetErrorOutput(w io.Writer) {
	errorOutput = &LockedWriter{writer: w, mutex: new(sync.Mutex)}
}

// GetErrorOutput returns the error output of the Print* functions
func GetErrorOutput() io.Writer {
	return errorOutput
}

// SetAllOutput changes the default and error output for the Print* functions
func SetAllOutput(w io.Writer) {
	single := new(sync.Mutex)
	terminalOutput = &LockedWriter{writer: w, mutex: single}
	errorOutput = &LockedWriter{writer: w, mutex: single}
}

// FlushAllOutput flushes all buffered output (if supported by the underlying Writer).
func FlushAllOutput() {
	for _, writer := range []io.Writer{terminalOutput, errorOutput} {
		if f, ok := writer.(Flusher); ok {
			_ = f.Flush()
		}
	}
}

// Print formats using the default formats for its operands and writes to standard output.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(terminalOutput, a...)
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(terminalOutput, a...)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(terminalOutput, format, a...)
}
