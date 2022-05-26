package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/monitor"
)

const (
	unixShell          = "sh"
	unixBashShell      = "bash"
	windowsShell       = "cmd.exe"
	windowsPowershell  = "powershell.exe"
	windowsPowershell6 = "pwsh.exe"
)

// SetPID is a callback to send the PID of the current child process
type SetPID func(pid int)

// ScanOutput is a callback to scan the default output of the command
// The implementation is expected to send everything read from the reader back to the writer
type ScanOutput func(r io.Reader, summary *monitor.Summary, w io.Writer) error

// Command holds the configuration to run a shell command
type Command struct {
	Command    string
	Arguments  []string
	Environ    []string
	Shell      []string
	Dir        string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	SetPID     SetPID
	ScanStdout ScanOutput
	sigChan    chan os.Signal
	done       chan interface{}
	analyser   *OutputAnalyser
}

// NewCommand instantiate a default Command without receiving OS signals (SIGTERM, etc.)
func NewCommand(command string, args []string) *Command {
	return &Command{
		Command:   command,
		Arguments: args,
		Environ:   []string{},
		analyser:  NewOutputAnalyser(),
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
		analyser:  NewOutputAnalyser(),
	}
}

// OnErrorCallback registers a custom callback that is invoked when pattern (regex) matches a line in stderr.
func (c *Command) OnErrorCallback(name, pattern string, minCount, maxCalls int, callback func(line string) error) error {
	return c.analyser.SetCallback(name, pattern, minCount, maxCalls, false, callback)
}

// Run the command
func (c *Command) Run() (monitor.Summary, string, error) {
	var err error
	var stdout, stderr io.ReadCloser

	summary := monitor.Summary{OutputAnalysis: c.analyser.Reset()}

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
		go func() {
			// send INT signal
			c.propagateSignal(cmd.Process)
			// close stdin (if possible) to unblock Wait on cmd.Process
			if in, canClose := cmd.Stdin.(io.Closer); canClose && in != nil {
				in.Close()
			}
		}()
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

		err = c.analyser.AnalyseLines(io.TeeReader(stderr, io.MultiWriter(stderrOutput, errors)))
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
	return summary, errorText, err
}

// getShellCommand transforms the command line and arguments to be launched via a shell (sh or cmd.exe)
func (c *Command) getShellCommand() (shell string, arguments []string, err error) {
	var searchList []string
	for _, sh := range c.Shell {
		if sh = strings.TrimSpace(sh); sh != "" {
			searchList = append(searchList, sh)
		}
	}
	if len(searchList) == 0 {
		searchList = c.getShellSearchList()
	}

	for _, search := range searchList {
		if shell, err = exec.LookPath(search); err == nil {
			break
		}
	}

	if err != nil {
		err = fmt.Errorf("cannot find shell (%s): %s", strings.Join(searchList, ", "), err.Error())
		return
	}

	arguments = c.composeShellArguments(shell)
	return
}

// removeQuotes removes single and double quotes when the whole string is quoted.
// this is only useful for windows where the arguments are sent one by one.
func removeQuotes(args []string) []string {
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
