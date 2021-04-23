package shell

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
)

const (
	windowsDefaultShell = "cmd.exe"
	windowsPowershell   = "powershell.exe"
)

// SetPID is a callback to send the PID of the current child process
type SetPID func(pid int)

// ScanOutput is a callback to scan the default output of the command
// The implementation is expected to send everything read from the reader back to the writer
type ScanOutput func(r io.Reader, summary *Summary, w io.Writer) error

// Command holds the configuration to run a shell command
type Command struct {
	Command       string
	Arguments     []string
	Environ       []string
	Dir           string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	SetPID        SetPID
	ScanStdout    ScanOutput
	UsePowershell bool
	sigChan       chan os.Signal
	done          chan interface{}
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
		done:      make(chan interface{}),
	}
}

// Run the command
func (c *Command) Run() (Summary, string, error) {
	var err error
	var stdout, stderr io.ReadCloser

	analyser := NewOutputAnalyser()
	summary := Summary{OutputAnalysis: analyser}

	command, args, err := c.getShellCommand()
	if err != nil {
		return summary, "", err
	}

	// clog.Tracef("command: %s %q", command, args)
	cmd := exec.Command(command, args...)

	if c.ScanStdout != nil {
		// install a pipe for scanning the output
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return summary, "", err
		}
	} else {
		cmd.Stdout = c.Stdout
	}
	if c.Stderr != nil {
		stderr, err = cmd.StderrPipe()
		if err != nil {
			cmd.Stderr = c.Stderr
		}
	}
	cmd.Stdin = c.Stdin

	cmd.Env = os.Environ()
	if c.Environ != nil && len(c.Environ) > 0 {
		cmd.Env = append(cmd.Env, c.Environ...)
	}

	start := time.Now()

	// spawn the child process
	if err = cmd.Start(); err != nil {
		return summary, "", err
	}
	if c.SetPID != nil {
		// send the PID back (to write down in a lockfile)
		c.SetPID(cmd.Process.Pid)
	}
	// setup the OS signalling if we need it (typically used for unixes but not windows)
	if c.sigChan != nil {
		defer func() {
			close(c.done)
		}()
		go c.propagateSignal(cmd.Process)
	}

	// output scanner
	if stdout != nil {
		err = c.ScanStdout(stdout, &summary, c.Stdout)
		if err != nil {
			return summary, "", err
		}
	}

	// handle command errors
	errors := &bytes.Buffer{}

	// send error output to buffer & stderr
	if stderr != nil {
		stderrOutput := c.Stderr
		if stderrOutput == nil {
			stderrOutput = os.Stderr
		}

		err = c.ScanStderr(stderr, stderrOutput, errors)
		if err != nil {
			clog.Errorf("failed reading stderr from command: %s ; Cause: %s", command, err.Error())
		}

		if e := cmd.Wait(); e != nil {
			err = e
		}
	} else {
		err = cmd.Wait()
	}

	// finish summary
	summary.Duration = time.Since(start)
	errorText := errors.String()
	analyser.AnalyseStringLines(errorText)

	return summary, errorText, err
}

// getShellCommand transforms the command line and arguments to be launched via a shell (sh or cmd.exe)
func (c *Command) getShellCommand() (string, []string, error) {

	if runtime.GOOS == "windows" {
		search := windowsDefaultShell
		if c.UsePowershell {
			search = windowsPowershell
		}
		shell, err := exec.LookPath(search)
		if err != nil {
			return "", nil, fmt.Errorf("cannot find shell executable (%s) in path", search)
		}
		// cmd.exe accepts that all arguments are sent one by one
		args := append([]string{"/C", c.Command}, removeQuotes(c.Arguments)...)
		return shell, args, nil
	}

	shell, err := exec.LookPath("sh")
	if err != nil {
		return "", nil, fmt.Errorf("cannot find shell executable (sh) in path")
	}
	// Flatten all arguments into one string, sh expects one big string
	flatCommand := append([]string{c.Command}, c.Arguments...)
	return shell, []string{"-c", strings.Join(flatCommand, " ")}, nil
}

func (c *Command) ScanStderr(r io.Reader, w1, w2 io.Writer) error {
	eol := []byte("\n")
	if runtime.GOOS == "windows" {
		eol = []byte("\r\n")
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		w1.Write(scanner.Bytes())
		w1.Write(eol)
		w2.Write(scanner.Bytes())
		w2.Write(eol)
	}

	if err := scanner.Err(); err != nil {
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
