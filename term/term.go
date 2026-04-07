package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/ansi"
	"github.com/mattn/go-colorable"
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
	go handleStatus()
	// must be last
	{
		setOutput(os.Stdout)
		setErrorOutput(os.Stderr)
	}
}

func handleStatus() {
	ticker := time.NewTicker(time.Second / StatusFPS)
	defer ticker.Stop()

	var waiting []chan bool
	respondWaiting := func(result bool) {
		for _, request := range waiting {
			request <- result
			close(request)
		}
		waiting = nil
	}
	defer respondWaiting(false)

	var newStatus, status []string
	buffer := &bytes.Buffer{}
	for {
		select {
		case lines := <-statusChannel:
			newStatus = lines

		case request := <-statusWaitChannel:
			waiting = append(waiting, request)

		case <-ticker.C:
			if status != nil && OutputIsTerminal() {
				width, height := OsStdoutTerminalSize()
				noAnsi := !IsColorableOutput()
				if height < 1 {
					continue
				} else if noAnsi {
					newStatus = newStatus[1:] // strip first empty line
					height = 1
				}
				if width >= 60 {
					width -= 2
				} else if width >= 80 {
					width -= 4 // right margin
				}

				last := truncate(status, height)
				printable := truncate(newStatus, height)
				removedLines := len(last) - len(printable)
				if removedLines > 0 {
					filler := make([]string, removedLines, removedLines+len(printable))
					printable = append(filler, printable...)
				}

				if len(printable) > 0 {
					buffer.Reset()
					for index, line := range printable {
						runes := []rune(strings.ReplaceAll(line, "\n", " "))
						_, maxIndex := ansi.RunesLength(runes, width)
						runes = truncate(runes, maxIndex)

						if noAnsi {
							if remaining := width - len(runes); remaining > 0 {
								for remaining > 0 {
									runes = append(runes, ' ')
									remaining--
								}
							}
							_, _ = fmt.Fprintf(buffer, "\r%s\r", string(runes))
						} else {
							eol := "\n"
							if index+1 == len(printable) {
								eol = "\r"
							}
							_, _ = fmt.Fprintf(buffer, "\r%s%s%s%s", ansi.ClearLine, string(runes), ansi.Reset, eol)
						}
					}

					if !noAnsi {
						buffer.WriteString(ansi.CursorUpLeftN(len(printable) - 1))
					}

					_, _ = buffer.WriteTo(getColorableOutput())
					buffer.Reset()
				}
			}
			status = newStatus
			respondWaiting(true)
		}
	}
}

func truncate[E any](src []E, maxLength int) []E {
	if len(src) > maxLength {
		return src[:maxLength]
	}
	return src
}

// SetStatus sets a status line(s) that is printed when the output is an interactive terminal
//
// Deprecated: use term.Terminal instead
func SetStatus(line []string) {
	// Clone lines and add empty line on top (= cursor position after printing status)
	if line != nil {
		line = append([]string{""}, line...)
	}
	statusChannel <- line
}

// WaitForStatus blocks until the previously provided status was applied
//
// Deprecated: use term.Terminal instead
func WaitForStatus() bool {
	request := make(chan bool, 1)
	statusWaitChannel <- request
	return <-request
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

// OsStdoutTerminalSize returns the width and height of the terminal session
//
// Deprecated: use term.Terminal instead
func OsStdoutTerminalSize() (width, height int) {
	fd := fdToInt(os.Stdout.Fd())
	var err error
	width, height, err = term.GetSize(fd)
	if err != nil {
		width, height = 0, 0
	}
	return
}

// OutputIsTerminal returns true if GetOutput sends to an interactive terminal
//
// Deprecated: use term.Terminal instead
func OutputIsTerminal() bool {
	return GetOutput() == os.Stdout && OsStdoutIsTerminal()
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
	SetStatus(nil)
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

// getColorableOutput returns an output supporting ANSI color if output is a terminal
func getColorableOutput() (out io.Writer) {
	if v := colorOutput.Load(); v != nil {
		out = *v
	}
	if out == nil {
		if IsColorableOutput() {
			out = colorable.NewColorable(os.Stdout)
		} else {
			out = colorable.NewNonColorable(outputWriter())
		}
		colorOutput.Store(&out)
	}
	return out
}

// IsColorableOutput tells whether GetColorableOutput supports ANSI color (and control characters) or discards ANSI
//
// Deprecated: use term.Terminal instead
func IsColorableOutput() bool {
	return enableColors.Load() && OutputIsTerminal()
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

// FlushAllOutput flushes all buffered output (if supported by the underlying Writer).
//
// Deprecated: use term.Terminal instead
func FlushAllOutput() {
	for _, writer := range []io.Writer{GetOutput(), GetErrorOutput()} {
		_, _ = util.FlushWriter(writer)
	}
}

func outputWriter() io.Writer {
	if PrintToError {
		return GetErrorOutput()
	}
	return GetOutput()
}
