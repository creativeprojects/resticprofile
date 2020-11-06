package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	terminalOutput io.Writer = os.Stdout
	errorOutput    io.Writer = os.Stderr
)

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

// SetOutput changes the default output for the Print* functions
func SetOutput(w io.Writer) {
	terminalOutput = w
}

// GetOutput returns the default output of the Print* functions
func GetOutput() io.Writer {
	return terminalOutput
}

// SetErrorOutput changes the error output for the Print* functions
func SetErrorOutput(w io.Writer) {
	errorOutput = w
}

// GetErrorOutput returns the error output of the Print* functions
func GetErrorOutput() io.Writer {
	return errorOutput
}

// SetAllOutput changes the default and error output for the Print* functions
func SetAllOutput(w io.Writer) {
	terminalOutput = w
	errorOutput = w
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
