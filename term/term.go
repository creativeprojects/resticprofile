package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/creativeprojects/resticprofile/util"
	"golang.org/x/term"
)

var (
	termOutput        atomic.Pointer[io.Writer]
	errorOutput       atomic.Pointer[io.Writer]
	colorOutput       atomic.Pointer[io.Writer]
	enableColors      atomic.Bool
	statusChannel     = make(chan []string)
	statusWaitChannel = make(chan chan bool)
	PrintToError      = false
)

const (
	StatusFPS = 10
)

func init() {
	enableColors.Store(true)
	// must be last
	{
		setOutput(os.Stdout)
		setErrorOutput(os.Stderr)
	}
}

func truncate[E any](src []E, maxLength int) []E {
	if len(src) > maxLength {
		return src[:maxLength]
	}
	return src
}

// ReadPassword reads a password without echoing it to the terminal.
//
// Deprecated: use term.Terminal instead
func ReadPassword() (string, error) {
	stdin := fdToInt(os.Stdin.Fd())
	if !term.IsTerminal(stdin) {
		return readLine()
	}
	line, err := term.ReadPassword(stdin)
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	return string(line), nil
}

// readLine reads some input
func readLine() (string, error) {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read line: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// OsStdoutIsTerminal returns true as os.Stdout is a terminal session
//
// Deprecated: use term.Terminal instead
func OsStdoutIsTerminal() bool {
	return isTerminal(os.Stdout)
}

// SetOutput changes the default output for the Print* functions
//
// Deprecated: use term.Terminal instead
func SetOutput(w io.Writer) {
	if w == os.Stdout && isTerminal(os.Stdout) {
		setOutput(os.Stdout)
	} else {
		setOutput(util.NewSyncWriter(w))
	}
}

func setOutput(w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	termOutput.Store(&w)
	colorOutput.Store(nil)
}

// GetOutput returns the default output of the Print* functions
//
// Deprecated: use term.Terminal instead
func GetOutput() (out io.Writer) {
	if v := termOutput.Load(); v != nil {
		out = *v
	}
	return
}

// SetErrorOutput changes the error output for the Print* functions
//
// Deprecated: use term.Terminal instead
func SetErrorOutput(w io.Writer) {
	if w == os.Stderr && isTerminal(os.Stderr) {
		setErrorOutput(os.Stderr)
	} else {
		setErrorOutput(util.NewSyncWriter(w))
	}
}

func setErrorOutput(w io.Writer) {
	if w == nil {
		w = io.Discard
	}
	errorOutput.Store(&w)
}

// GetErrorOutput returns the error output of the Print* functions
//
// Deprecated: use term.Terminal instead
func GetErrorOutput() (out io.Writer) {
	if v := errorOutput.Load(); v != nil {
		out = *v
	}
	return
}

// SetAllOutput changes the default and error output for the Print* functions
//
// Deprecated: use term.Terminal instead
func SetAllOutput(w io.Writer) {
	SetOutput(w)
	setErrorOutput(GetOutput())
}
