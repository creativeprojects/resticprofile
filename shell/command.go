package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Command holds the configuration to run a shell command
type Command struct {
	Command   string
	Arguments []string
	Environ   []string
	Dir       string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	sigChan   chan os.Signal
	done      chan bool
}

// NewCommand instantiate a default Command without receiving OS signals (SIGTERM, etc.)
func NewCommand(command string, args []string) *Command {
	return &Command{
		Command:   command,
		Arguments: args,
		Environ:   []string{},
	}
}

// NewSignalledCommand instantiate a default Command receiving OS signals (SIGTERM, etc.)
func NewSignalledCommand(command string, args []string, c chan os.Signal) *Command {
	return &Command{
		Command:   command,
		Arguments: args,
		Environ:   []string{},
		sigChan:   c,
		done:      make(chan bool, 1),
	}
}

// Run the command
func (c *Command) Run() error {
	var err error

	command, args, err := getShellCommand(c.Command, c.Arguments)
	if err != nil {
		return err
	}

	cmd := exec.Command(command, args...)

	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	cmd.Stdin = c.Stdin

	cmd.Env = os.Environ()
	if c.Environ != nil && len(c.Environ) > 0 {
		cmd.Env = append(cmd.Env, c.Environ...)
	}

	if c.sigChan == nil {
		// Simple case without taking care of OS signals
		err = cmd.Run()
		if err != nil {
			return err
		}
		return nil
	}

	// More complex case with OS signalling
	if err = cmd.Start(); err != nil {
		return err
	}
	defer func() {
		close(c.done)
	}()
	go c.propagateSignal(cmd.Process)
	return cmd.Wait()
}

// getShellCommand transforms the command line and arguments to be launched via a shell (sh or cmd.exe)
func getShellCommand(command string, args []string) (string, []string, error) {

	if runtime.GOOS == "windows" {
		shell, err := exec.LookPath("cmd.exe")
		if err != nil {
			return "", nil, fmt.Errorf("cannot find shell executable (cmd.exe) in path")
		}
		// cmd.exe accepts that all arguments are sent one bye one
		args := append([]string{"/C", command}, removeQuotes(args)...)
		return shell, args, nil
	}

	shell, err := exec.LookPath("sh")
	if err != nil {
		return "", nil, fmt.Errorf("cannot find shell executable (sh) in path")
	}
	// Flatten all arguments into one string, sh expects one big string
	flatCommand := append([]string{command}, args...)
	return shell, []string{"-c", strings.Join(flatCommand, " ")}, nil
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
