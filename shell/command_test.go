package shell

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockBinary string
)

func init() {
	// all the tests are running in the exec directory
	if runtime.GOOS == "windows" {
		mockBinary = "..\\mock.exe"
	} else {
		mockBinary = "../mock"
	}
	// build restic mock
	cmd := exec.Command("go", "build", "-o", mockBinary, "./mock")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("cannot build mock: %q", string(output)))
	}
}

func TestRemoveQuotes(t *testing.T) {
	source := []string{
		`-p`,
		`"file"`,
		`--other`,
		`'file'`,
		`Cote d'ivoire`,
	}
	unquoted := removeQuotes(source)
	assert.Equal(t, []string{
		"-p",
		"file",
		"--other",
		"file",
		"Cote d'ivoire",
	}, unquoted)
}

func TestShellCommandWithArguments(t *testing.T) {
	testCommand := "/bin/restic"
	testArgs := []string{
		`-v`,
		`--exclude-file`,
		`"excludes"`,
		`--repo`,
		`"/path/with space"`,
		`backup`,
		`.`,
	}
	c := &Command{
		Command:   testCommand,
		Arguments: testArgs,
	}

	command, args, err := c.GetShellCommand()
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			`/V:ON`, `/C`,
			`/bin/restic`,
			`-v`,
			`--exclude-file`,
			`excludes`,
			`--repo`,
			`/path/with space`,
			`backup`,
			`.`,
		}, args)
	} else {
		assert.Regexp(t, regexp.MustCompile("(/usr)?/bin/(ba)?sh"), command)
		assert.Equal(t, []string{
			"-c",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	}
}

func TestShellCommand(t *testing.T) {
	testCommand := "/bin/restic -v --exclude-file \"excludes\" --repo \"/path/with space\" backup ."
	testArgs := []string{}
	c := &Command{
		Command:   testCommand,
		Arguments: testArgs,
	}

	command, args, err := c.GetShellCommand()
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			"/V:ON", "/C",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	} else {
		assert.Regexp(t, regexp.MustCompile("(/usr)?/bin/(ba)?sh"), command)
		assert.Equal(t, []string{
			"-c",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	}
}

func TestShellArgumentsComposing(t *testing.T) {
	tests := []struct {
		command               string
		shell, args, expected []string
	}{
		{
			shell:    []string{defaultShell, bashShell},
			command:  mockBinary,
			args:     []string{"a", "\"-$PROFILE_NAME- -\"", "c"},
			expected: []string{"-c", mockBinary + " a \"-$PROFILE_NAME- -\" c"},
		},
		{
			shell:    []string{windowsShell},
			command:  mockBinary,
			args:     []string{"a", "\"-$PROFILE_NAME- -\"", "c", "!custom! %custom% %PATH%"},
			expected: []string{"/V:ON", "/C", mockBinary, "a", "-!PROFILE_NAME!- -", "c", "!custom! %custom% %PATH%"},
		},
		{
			shell:    []string{windowsShell},
			command:  "echo \"$PROFILE_NAME\"",
			args:     nil,
			expected: []string{"/V:ON", "/C", "echo \"!PROFILE_NAME!\""},
		},
		{
			shell:   []string{powershell, powershell6},
			command: mockBinary,
			args:    []string{"a", "\"-$PROFILE_NAME- -\"", "$True $Env:custom ${Env:c2} $$ $? $Error \"$home\""},
			expected: []string{
				"-Command", mockBinary, "a",
				"-$($_=${Env:PROFILE_NAME}; if ($_) {$_} else {${PROFILE_NAME}})- -",
				"${True} $Env:custom ${Env:c2} $$ $? ${Error} \"${home}\"",
			},
		},
		{
			shell:    []string{powershell, powershell6},
			command:  "echo \"$PROFILE_NAME\"",
			args:     nil,
			expected: []string{"-Command", "echo \"$($_=${Env:PROFILE_NAME}; if ($_) {$_} else {${PROFILE_NAME}})\""},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			var shells []string
			for _, shell := range test.shell {
				shells = append(shells, shell)
				shells = append(shells, shell+".exe")
				shells = append(shells, strings.ToUpper(shell+".exe"))
				shells = append(shells, "some/path/to/"+shell)
			}
			for _, shell := range shells {
				t.Run(shell, func(t *testing.T) {
					c := &Command{Shell: []string{shell}, Command: test.command, Arguments: test.args}
					composedArgs := getArgumentsComposer(shell)(c)
					assert.Equal(t, test.expected, composedArgs)
				})
			}
		})
	}
}

func TestVariablesRewrite(t *testing.T) {
	mapper := func(name string) string { return fmt.Sprintf("!%s!", name) }

	tests := []struct{ in, expected string }{
		{in: "arg", expected: "arg"},
		{in: "no vars", expected: "no vars"},
		{in: "no vars: %no_unix%", expected: "no vars: %no_unix%"},
		{in: "no vars: $-no- $(123) $$ $? ${a:-1} ${a:-1}", expected: "no vars: $-no- $(123) $$ $? ${a:-1} ${a:-1}"},
		{in: "no vars: $Env:abc ${Env:abc}", expected: "no vars: $Env:abc ${Env:abc}"},
		{in: "$a", expected: "!a!"},
		{in: "${a}", expected: "!a!"},
		{in: "${_a_}", expected: "!_a_!"},
		{in: "$_a_", expected: "!_a_!"},
		{in: "$a ${b}_$c_-$d", expected: "!a! !b!_!c_!-!d!"},
		{in: "$d$x$y", expected: "!d!$x!y!"}, // known issue, won't fix
		{in: "${d}${x}${y}", expected: "!d!!x!!y!"},
	}

	frame := []string{"", "-", "[", "]", "-token", " token", " token ", " _ "}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			for _, prefix := range frame {
				for _, suffix := range frame {
					input := prefix + test.in + suffix
					expected := prefix + test.expected + suffix

					actual := rewriteVariables([]string{input}, mapper)[0]
					assert.Equal(t, expected, actual)
				}
			}
		})
	}
}

func TestRunLocalCommand(t *testing.T) {
	cmd := NewCommand("command.go", nil)
	expected := "." + string(os.PathSeparator) + "command.go"

	for _, shell := range []string{defaultShell, bashShell, powershell, powershell6, windowsShell} {
		args := getArgumentsComposer(shell)(cmd)
		assert.Equal(t, expected, args[len(args)-1])
	}
}

func TestShellSearchPath(t *testing.T) {
	searchList := NewCommand("echo", []string{}).getShellSearchList()
	assert.NotEmpty(t, searchList)
	for _, shell := range searchList {
		assert.NotNil(t, shellArgumentsComposerRegistry[shell])
	}
}

func TestSelectCustomShell(t *testing.T) {
	cmd := NewCommand("sleep", []string{"1"})
	cmd.Shell = []string{mockBinary}
	shell, _, err := cmd.GetShellCommand()
	assert.Nil(t, err)
	assert.Equal(t, mockBinary, shell)

	expected := "cannot find shell: exec: \"non-existing-shell\": executable file not found in $PATH (tried non-existing-shell)"
	if runtime.GOOS == "windows" {
		expected = strings.ReplaceAll(expected, "$PATH", "%PATH%")
	}

	cmd.Shell = []string{"non-existing-shell"}
	shell, _, err = cmd.GetShellCommand()
	assert.EqualError(t, err, expected)
	assert.Empty(t, shell)
}

func TestRunShellEcho(t *testing.T) {
	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{"TestRunShellEcho"})
	cmd.Stdout = buffer
	_, _, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, string(output), "TestRunShellEcho")
}

func TestRunShellEchoWithSignalling(t *testing.T) {
	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := NewSignalledCommand("echo", []string{"TestRunShellEchoWithSignalling"}, c)
	cmd.Stdout = buffer
	_, _, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, string(output), "TestRunShellEchoWithSignalling")
}

// There is something wrong with this test under Linux
func TestInterruptShellCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test not running on this platform")
	}
	if runtime.GOOS == "linux" {
		t.Skip("Test not running on this platform")
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)

	cmd := NewSignalledCommand(mockBinary, []string{"test", "--sleep", "3000"}, sigChan)
	cmd.Stdout = buffer

	// Will ask us to stop in 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- syscall.Signal(syscall.SIGINT)
	}()
	start := time.Now()
	_, _, err := cmd.Run()
	// GitHub Actions *sometimes* sends a different message: "signal: interrupt"
	if err != nil && err.Error() != "exit status 128" && err.Error() != "signal: interrupt" {
		t.Fatal(err)
	}

	// check it ran for more than 100ms (but less than 300ms - the build agent can be very slow at times)
	duration := time.Since(start)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
	assert.Less(t, duration.Milliseconds(), int64(300))
}

func TestSetPIDCallback(t *testing.T) {
	called := 0
	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{"TestSetPIDCallback"})
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int) {
		called++
	}
	_, _, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, called)
}

func TestSetPIDCallbackWithSignalling(t *testing.T) {
	called := 0
	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := NewSignalledCommand("echo", []string{"TestSetPIDCallbackWithSignalling"}, c)
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int) {
		called++
	}
	_, _, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, called)
}

func TestSummaryDurationCommand(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	cmd := NewCommand("sleep", []string{"1"})
	if runtime.GOOS == "windows" {
		cmd.Shell = []string{powershell}
	}
	cmd.Stdout = buffer

	start := time.Now()
	summary, _, err := cmd.Run()
	require.NoError(t, err)

	// make sure the command ran properly
	assert.WithinDuration(t, time.Now(), start.Add(1*time.Second), 500*time.Millisecond)
	assert.GreaterOrEqual(t, summary.Duration.Milliseconds(), int64(1000))
	assert.Less(t, summary.Duration.Milliseconds(), int64(1500))
}

func TestSummaryDurationSignalledCommand(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)
	cmd := NewSignalledCommand("sleep", []string{"1"}, sigChan)
	if runtime.GOOS == "windows" {
		cmd.Shell = []string{powershell}
	}
	cmd.Stdout = buffer

	start := time.Now()
	summary, _, err := cmd.Run()
	require.NoError(t, err)

	// make sure the command ran properly
	assert.WithinDuration(t, time.Now(), start.Add(1*time.Second), 500*time.Millisecond)
	assert.GreaterOrEqual(t, summary.Duration.Milliseconds(), int64(1000))
	assert.Less(t, summary.Duration.Milliseconds(), int64(1500))
}

func TestStderr(t *testing.T) {
	expected := "error message\n"
	if runtime.GOOS == "windows" {
		expected = "\"error message\" \r\n"
	}

	cmd := NewCommand("echo", []string{"error message", ">&2"})
	bufferStdout, bufferStderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout = bufferStdout
	cmd.Stderr = bufferStderr
	_, stderr, err := cmd.Run()
	require.NoError(t, err)
	assert.Empty(t, bufferStdout.String())
	assert.Equal(t, expected, stderr)
}

func TestStderrSignalledCommand(t *testing.T) {
	expected := "error message\n"
	if runtime.GOOS == "windows" {
		expected = "\"error message\" \r\n"
	}

	sigChan := make(chan os.Signal, 1)
	cmd := NewSignalledCommand("echo", []string{"error message", ">&2"}, sigChan)
	bufferStdout, bufferStderr := &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout = bufferStdout
	cmd.Stderr = bufferStderr
	_, stderr, err := cmd.Run()
	require.NoError(t, err)
	assert.Empty(t, bufferStdout.String())
	assert.Equal(t, expected, stderr)
}

func TestStderrNotRedirected(t *testing.T) {
	cmd := NewCommand("echo", []string{"error message", ">&2"})
	bufferStdout := &bytes.Buffer{}
	cmd.Stdout = bufferStdout
	cmd.Stderr = nil
	_, stderr, err := cmd.Run()
	require.NoError(t, err)
	assert.Empty(t, bufferStdout.String())
	assert.Equal(t, "", stderr)
}

func TestStderrNotRedirectedSignalledCommand(t *testing.T) {
	sigChan := make(chan os.Signal, 1)
	cmd := NewSignalledCommand("echo", []string{"error message", ">&2"}, sigChan)
	bufferStdout := &bytes.Buffer{}
	cmd.Stdout = bufferStdout
	cmd.Stderr = nil
	_, stderr, err := cmd.Run()
	require.NoError(t, err)
	assert.Empty(t, bufferStdout.String())
	assert.Equal(t, "", stderr)
}

// Try to make a test to make sure restic is catching the signal properly
func TestResticCanCatchInterruptSignal(t *testing.T) {
	// if runtime.GOOS == "windows" {
	// 	t.Skip("cannot send a signal to a child process in Windows")
	// }

	// childPID := 0
	// var err error
	// buffer := &bytes.Buffer{}
	// cmd := NewCommand("restic", []string{"version"})
	// cmd.SetPID = func(pid int) {
	// 	childPID = pid
	// 	t.Logf("child PID = %d", childPID)
	// 	// release the current goroutine
	// 	go func(t *testing.T, pid int) {
	// 		time.Sleep(1 * time.Millisecond)
	// 		process, err := os.FindProcess(pid)
	// 		require.NoError(t, err)
	// 		t.Log("send SIGINT")
	// 		err = process.Signal(syscall.SIGINT)
	// 		require.NoError(t, err)
	// 	}(t, childPID)
	// }
	// cmd.Stdout = buffer
	// _, err = cmd.Run()
	// assert.Error(t, err)
}

func TestCanAnalyseLockFailure(t *testing.T) {
	file, err := ioutil.TempFile(".", "test-restic-lock-failure")
	require.NoError(t, err)
	file.Write([]byte(ResticLockFailureOutput))
	file.Close()
	fileName := file.Name()
	defer os.Remove(fileName)

	cmd := NewCommand(mockBinary, []string{"test", "--stderr", fmt.Sprintf("@%s", fileName)})
	cmd.Stderr = &bytes.Buffer{}

	summary, _, err := cmd.Run()
	assert.NoError(t, err)
	assert.NotNil(t, summary.OutputAnalysis)
	assert.True(t, summary.OutputAnalysis.ContainsRemoteLockFailure())
}

func TestCanAnalyseWithCustomPattern(t *testing.T) {
	reportedLine := ""
	expectedLine := "--content-to-match--"
	cmd := NewCommand(mockBinary, []string{"test", "--stderr", expectedLine})
	cmd.Stderr = &bytes.Buffer{}

	err := cmd.OnErrorCallback("cb-test", ".*content-to-match.*", 0, 0, func(line string) error {
		reportedLine = line
		return nil
	})
	assert.Nil(t, err)

	_, _, err = cmd.Run()
	assert.NoError(t, err)
	assert.Equal(t, expectedLine, reportedLine)
}
