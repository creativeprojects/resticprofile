package term

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/ansi"
	"github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	termOutput        atomic.Pointer[io.Writer]
	errorOutput       atomic.Pointer[io.Writer]
	colorOutput       atomic.Pointer[io.Writer]
	enableColors      atomic.Bool
	statusChannel     = make(chan []string)
	statusWaitChannel = make(chan chan bool)
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

					_, _ = buffer.WriteTo(GetColorableOutput())
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
func SetStatus(line []string) {
	// Clone lines and add empty line on top (= cursor position after printing status)
	if line != nil {
		line = append([]string{""}, line...)
	}
	statusChannel <- line
}

// WaitForStatus blocks until the previously provided status was applied
func WaitForStatus() bool {
	request := make(chan bool, 1)
	statusWaitChannel <- request
	return <-request
}

// AskYesNo prompts the user for a message asking for a yes/no answer
func AskYesNo(reader io.Reader, message string, defaultAnswer bool) bool {
	if !strings.HasSuffix(message, "?") {
		message += "?"
	}
	question := ""
	input := ""
	if defaultAnswer {
		question = "(Y/n)"
		input = "y"
	} else {
		question = "(y/N)"
		input = "n"
	}
	_, _ = Printf("%s %s: ", message, question)
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
	if !terminal.IsTerminal(stdin) {
		return ReadLine()
	}
	line, err := terminal.ReadPassword(stdin)
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	return string(line), nil
}

// ReadLine reads some input
func ReadLine() (string, error) {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read line: %v", err)
	}
	return strings.TrimSpace(line), nil
}

// OsStdoutIsTerminal returns true as os.Stdout is a terminal session
func OsStdoutIsTerminal() bool {
	return isTerminal(os.Stdout)
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	fd := int(file.Fd())
	return terminal.IsTerminal(fd)
}

// OsStdoutTerminalSize returns the width and height of the terminal session
func OsStdoutTerminalSize() (width, height int) {
	fd := int(os.Stdout.Fd())
	var err error
	width, height, err = terminal.GetSize(fd)
	if err != nil {
		width, height = 0, 0
	}
	return
}

// OutputIsTerminal returns true if GetOutput sends to an interactive terminal
func OutputIsTerminal() bool {
	return GetOutput() == os.Stdout && OsStdoutIsTerminal()
}

// SetOutput changes the default output for the Print* functions
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
	termOutput.Store(util.CopyRef[io.Writer](w))
	colorOutput.Store(nil)
	SetStatus(nil)
}

// GetOutput returns the default output of the Print* functions
func GetOutput() (out io.Writer) {
	if v := termOutput.Load(); v != nil {
		out = *v
	}
	return
}

// GetColorableOutput returns an output supporting ANSI color if output is a terminal
func GetColorableOutput() (out io.Writer) {
	if v := colorOutput.Load(); v != nil {
		out = *v
	}
	if out == nil {
		if IsColorableOutput() {
			out = colorable.NewColorable(os.Stdout)
		} else {
			out = colorable.NewNonColorable(GetOutput())
		}
		colorOutput.Store(util.CopyRef[io.Writer](out))
	}
	return out
}

// EnableColorableOutput toggles whether GetColorableOutput supports ANSI color or discards ANSI
func EnableColorableOutput(enable bool) {
	if enableColors.CompareAndSwap(!enable, enable) {
		colorOutput.Store(nil)
	}
}

// IsColorableOutput tells whether GetColorableOutput supports ANSI color (and control characters) or discards ANSI
func IsColorableOutput() bool {
	return enableColors.Load() && OutputIsTerminal()
}

// SetErrorOutput changes the error output for the Print* functions
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
	errorOutput.Store(util.CopyRef[io.Writer](w))
}

// GetErrorOutput returns the error output of the Print* functions
func GetErrorOutput() (out io.Writer) {
	if v := errorOutput.Load(); v != nil {
		out = *v
	}
	return
}

// SetAllOutput changes the default and error output for the Print* functions
func SetAllOutput(w io.Writer) {
	SetOutput(w)
	setErrorOutput(GetOutput())
}

// FlushAllOutput flushes all buffered output (if supported by the underlying Writer).
func FlushAllOutput() {
	for _, writer := range []io.Writer{GetOutput(), GetErrorOutput()} {
		_, _ = util.FlushWriter(writer)
	}
}

var recording = struct {
	lock          sync.Mutex
	buffer        *bytes.Buffer
	writer        io.Writer
	output, error io.Writer
}{}

type RecordMode uint8

const (
	RecordOutput RecordMode = iota
	RecordError
	RecordBoth
)

func StartRecording(mode RecordMode) {
	recording.lock.Lock()
	defer recording.lock.Unlock()
	if recording.buffer != nil {
		return
	}

	recording.buffer = new(bytes.Buffer)
	recording.writer = util.NewSyncWriterMutex(recording.buffer, &recording.lock)

	if mode != RecordError {
		recording.output = GetOutput()
		setOutput(recording.writer)
	}
	if mode != RecordOutput {
		recording.error = GetErrorOutput()
		setErrorOutput(recording.writer)
	}
}

func ReadRecording() (content string) {
	recording.lock.Lock()
	defer recording.lock.Unlock()
	if recording.buffer != nil {
		content = recording.buffer.String()
		recording.buffer.Reset()
	}
	return
}

func StopRecording() (content string) {
	recording.lock.Lock()
	defer recording.lock.Unlock()
	if recording.buffer != nil {
		if recording.output != nil && recording.writer == GetOutput() {
			setOutput(recording.output)
			recording.output = nil
		}

		if recording.error != nil && recording.writer == GetErrorOutput() {
			setErrorOutput(recording.error)
			recording.error = nil
		}

		content = recording.buffer.String()
		recording.writer = nil
		recording.buffer = nil
	}
	return
}

// Print formats using the default formats for its operands and writes to standard output.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(GetColorableOutput(), a...)
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(GetColorableOutput(), a...)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(GetColorableOutput(), format, a...)
}
