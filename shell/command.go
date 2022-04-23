package shell

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/monitor"
)

const (
	defaultShell = "sh"
	bashShell    = "bash"
	powershell   = "powershell"
	powershell6  = "pwsh"
	windowsShell = "cmd"
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

	command, args, err := c.GetShellCommand()
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

// GetShellCommand transforms the command line and arguments to be launched via a shell (sh or cmd.exe)
func (c *Command) GetShellCommand() (shell string, arguments []string, err error) {
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
		err = fmt.Errorf("cannot find shell: %w (tried %s)", err, strings.Join(searchList, ", "))
		return
	}

	composer := getArgumentsComposer(shell)
	arguments = composer(c)
	return
}

type shellArgumentsComposer func(*Command) []string

var shellArgumentsComposerRegistry map[string]shellArgumentsComposer

func init() {
	shellArgumentsComposerRegistry = map[string]shellArgumentsComposer{
		defaultShell: composeShellArguments,
		bashShell:    composeShellArguments,
		powershell:   composePowershellArguments,
		powershell6:  composePowershellArguments,
		windowsShell: composeWindowsCmdArguments,
	}
}

func getArgumentsComposer(shell string) (composer shellArgumentsComposer) {
	shell = strings.ToLower(filepath.Base(shell))

	if ext := filepath.Ext(shell); len(ext) > 0 {
		shell = shell[:len(shell)-len(ext)]
	}

	composer = shellArgumentsComposerRegistry[shell]
	if composer == nil {
		composer = shellArgumentsComposerRegistry[defaultShell]
	}
	return
}

func composeShellArguments(c *Command) []string {
	// Flatten all arguments into one string, sh and bash expects one big string
	command := resolveCommand(c.Command)
	flatCommand := strings.Join(append([]string{command}, c.Arguments...), " ")

	return []string{
		"-c",
		flatCommand,
	}
}

var powershellBuiltins = regexp.MustCompile("(?i)^(\\$|\\?|^|_|args|ConsoleFileName|Error|ErrorView|" +
	"Event|EventArgs|EventSubscriber|ExecutionContext|false|foreach|HOME|Host|input|IsCoreCLR|" +
	"IsLinux|IsMacOS|IsWindows|LastExitCode|Matches|MyInvocation|NestedPromptLevel|null|PID|PROFILE|" +
	"PSBoundParameters|PSCmdlet|PSCommandPath|PSCulture|PSDebugContext|PSHOME|PSItem|" +
	"PSNativeCommandArgumentPassing|PSScriptRoot|PSSenderInfo|PSStyle|PSUICulture|PSVersionTable|" +
	"PWD|Sender|ShellId|StackTrace|switch|this|True)$")

func composePowershellArguments(c *Command) []string {
	// Rewrite unix style env variables ($var) to powershell env style ($Env:var) with fallback to local variable
	mapper := func(name string) string {
		if powershellBuiltins.MatchString(name) {
			return fmt.Sprintf("${%s}", name)
		} else {
			return fmt.Sprintf("${Env:%s}", name)
		}
	}
	arguments := rewriteVariables(c.Arguments, mapper)
	command := resolveCommand(rewriteVariables([]string{c.Command}, mapper)[0])

	return append(
		[]string{"-Command", command},
		removeQuotes(arguments)...,
	)
}

func composeWindowsCmdArguments(c *Command) []string {
	// Rewrite unix style env variables ($var) to delayed cmd style (!var!)
	mapper := func(name string) string { return fmt.Sprintf("!%s!", name) }
	arguments := rewriteVariables(c.Arguments, mapper)
	command := resolveCommand(rewriteVariables([]string{c.Command}, mapper)[0])

	// Enable delayed variable expansion "/V:ON" to support !variable! syntax
	return append(
		[]string{"/V:ON", "/C", command},
		removeQuotes(arguments)...,
	)
}

// resolveCommand adds a "./" prefix to a command that only exists in the current working directory
func resolveCommand(command string) string {
	// Check if a command was specified that only exists in CWD (without "./" prefix)
	// this happens for example if the restic binary is placed in CWD
	if !strings.HasPrefix(command, ".") &&
		!strings.Contains(command, " ") &&
		filepath.Base(command) == command {
		if cmd, err := exec.LookPath(command); errors.Is(err, exec.ErrNotFound) || cmd == command {
			if s, err := os.Stat(command); err == nil && !s.IsDir() && s.Size() > 0 {
				command = "." + string(os.PathSeparator) + command
			}
		}
	}

	return command
}

// unixVariablesMatcher matches all "$var" and "${var}" inside a string (excluding "$var.prop", "$var[]" & "$var:prop")
var unixVariablesMatcher = regexp.MustCompile(`(?i)\$(\w+)([^\w:.\[]|$)|\$\{(\w+)}`)

// makeRegexVariablesFunc converts a variables expand func to a func for ReplaceAllStringFunc
func makeRegexVariablesFunc(mapper func(string) string) func(string) string {
	return func(fullMatch string) (mapped string) {
		if matches := unixVariablesMatcher.FindStringSubmatch(fullMatch); len(matches) > 1 {
			for _, name := range matches[1:] {
				if len(name) > 0 {
					if len(mapped) == 0 {
						mapped = mapper(name)
					} else {
						mapped += name
					}
				}
			}
		}
		return
	}
}

// rewriteVariables iterates arguments and formats every unix style variable into a new result slice
func rewriteVariables(arguments []string, mapper func(string) string) (result []string) {
	regexMapper := makeRegexVariablesFunc(mapper)
	for _, arg := range arguments {
		arg = unixVariablesMatcher.ReplaceAllStringFunc(arg, regexMapper)
		result = append(result, arg)
	}
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
