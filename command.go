package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type commandDefinition struct {
	command       string
	args          []string
	env           []string
	displayStderr bool
	useStdin      bool
	shell         bool
}

func newCommand(command string, args, env []string) commandDefinition {
	if env == nil {
		env = make([]string, 0)
	}
	return commandDefinition{
		command:       command,
		args:          args,
		env:           env,
		displayStderr: true,
		useStdin:      false,
		shell:         true,
	}
}

func runCommands(commands []commandDefinition) error {
	for _, command := range commands {
		err := runCommand(command)
		if err != nil {
			return err
		}
	}
	return nil
}

func runCommand(command commandDefinition) error {
	var cmd *exec.Cmd
	var err error

	if !command.shell {
		cmd = exec.Command(command.command, command.args...)
	} else {
		cmd, err = getShellCommand(command.command, command.args)
		if err != nil {
			return err
		}
	}

	cmd.Stdout = os.Stdout
	if command.displayStderr {
		cmd.Stderr = os.Stderr
	}

	if command.useStdin {
		cmd.Stdin = os.Stdin
	}

	cmd.Env = os.Environ()
	if command.env != nil && len(command.env) > 0 {
		cmd.Env = append(cmd.Env, command.env...)
	}

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// getShellCommand transforms the command line and arguments to be launched via a shell (sh or cmd.exe)
func getShellCommand(command string, args []string) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		shell, err := exec.LookPath("cmd.exe")
		if err != nil {
			return nil, fmt.Errorf("Cannot find shell executable (cmd.exe) in path")
		}
		args := append([]string{"/C", command}, removeQuotes(args)...)
		cmd = exec.Command(shell, args...)

	} else {
		shell, err := exec.LookPath("sh")
		if err != nil {
			return nil, fmt.Errorf("Cannot find shell executable (sh) in path")
		}
		flatCommand := append([]string{command}, args...)
		cmd = exec.Command(shell, "-c", strings.Join(flatCommand, " "))
	}
	return cmd, nil
}

func runShellCommand(command string) error {
	var cmd *exec.Cmd
	var err error

	if runtime.GOOS == "windows" {
		shell, err := exec.LookPath("cmd.exe")
		if err != nil {
			return fmt.Errorf("Cannot find shell executable (sh) in path")
		}
		cmd = exec.Command(shell, "/C", command)

	} else {
		shell, err := exec.LookPath("sh")
		if err != nil {
			return fmt.Errorf("Cannot find shell executable (sh) in path")
		}
		cmd = exec.Command(shell, "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// removeQuotes removes single and double quotes when the whole string is quoted
func removeQuotes(args []string) []string {
	if args == nil {
		return nil
	}

	singleQuote := `'`
	doubleQuote := `"`

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], doubleQuote) && strings.HasSuffix(args[i], doubleQuote) {
			args[i] = strings.Trim(args[i], doubleQuote)

		} else if strings.HasPrefix(args[i], singleQuote) && strings.HasSuffix(args[i], singleQuote) {
			args[i] = strings.Trim(args[i], singleQuote)
		}
	}
	return args
}
