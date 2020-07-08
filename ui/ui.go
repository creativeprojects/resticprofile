package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
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
		return "", fmt.Errorf("Failed to read password: %v", err)
	}
	return string(line), nil
}

// ReadLine reads some input
func ReadLine() (string, error) {
	buf := bufio.NewReader(os.Stdin)
	line, err := buf.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("Failed to read line: %v", err)
	}
	return strings.TrimSpace(line), nil
}
