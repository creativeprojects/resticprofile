package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockBinary string
)

func TestMain(m *testing.M) {
	// using an anonymous function to handle defer statements before os.Exit()
	exitCode := func() int {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		tempDir, err := os.MkdirTemp("", "resticprofile-shell")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot create temp dir: %v\n", err)
			return 1
		}
		fmt.Printf("using temporary dir: %q\n", tempDir)
		defer os.RemoveAll(tempDir)

		mockBinary = filepath.Join(tempDir, "shellmock")
		if platform.IsWindows() {
			mockBinary += ".exe"
		}
		cmd := exec.CommandContext(ctx, "go", "build", "-buildvcs=false", "-o", mockBinary, "./mock")
		if output, err := cmd.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "Error building mock binary: %s\nCommand output: %s\n", err, string(output))
			return 1
		}
		return m.Run()
	}()
	os.Exit(exitCode)
}

func TestRemoveQuotes(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	testCommand := `"/bin/with space/restic"`
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
	require.NoError(t, err)
	if platform.IsWindows() {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			`/V:ON`, `/C`,
			`"/bin/with space/restic"`,
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
			"\"/bin/with space/restic\" -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	}
}

func TestShellCommand(t *testing.T) {
	t.Parallel()

	testCommand := "\"/bin/with space/restic\" -v --exclude-file \"excludes\" --repo \"/path/with space\" backup ."
	testArgs := []string{}
	c := &Command{
		Command:   testCommand,
		Arguments: testArgs,
	}

	command, args, err := c.GetShellCommand()
	require.NoError(t, err)
	if platform.IsWindows() {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			"/V:ON", "/C",
			"\"/bin/with space/restic\" -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	} else {
		assert.Regexp(t, regexp.MustCompile("(/usr)?/bin/(ba)?sh"), command)
		assert.Equal(t, []string{
			"-c",
			"\"/bin/with space/restic\" -v --exclude-file \"excludes\" --repo \"/path/with space\" backup .",
		}, args)
	}
}

func TestShellArgumentsComposing(t *testing.T) {
	t.Parallel()

	exeWithSpace := filepath.Join(t.TempDir(), "some folder", "executable")
	require.NoError(t, os.MkdirAll(filepath.Dir(exeWithSpace), 0o700))
	require.NoError(t, os.WriteFile(exeWithSpace, []byte{0}, 0o600))

	tests := []struct {
		command               string
		shell, args, expected []string
	}{
		{
			shell:    []string{defaultShell, bashShell, "any-other-shell"},
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
			args:    []string{"a", "\"-$PROFILE_NAME- -\"", "$True $Env:custom $custom ${Env:c2} $$ $? $Error \"${home}\""},
			expected: []string{
				"-Command", mockBinary, "\"a\"",
				"\"-${Env:PROFILE_NAME}- -\"",
				"\"$True $Env:custom $custom ${Env:c2} $$ $? $Error `\"${home}`\"\"",
			},
		},
		{
			shell:    []string{powershell, powershell6},
			command:  "echo \"$PROFILE_NAME\"",
			args:     nil,
			expected: []string{"-Command", "echo \"${Env:PROFILE_NAME}\""},
		},
		{
			shell:    []string{powershell, powershell6},
			command:  `$PROFILE_NAME = "custom"; echo "$PROFILE_NAME"`,
			args:     nil,
			expected: []string{"-Command", `$PROFILE_NAME = "custom"; echo "$PROFILE_NAME"`},
		},
		{
			shell:    []string{powershell, powershell6},
			command:  `New-Variable -name "Profile_Name" && echo "$PROFILE_NAME"`,
			args:     nil,
			expected: []string{"-Command", `New-Variable -name "Profile_Name" && echo "$PROFILE_NAME"`},
		},
		{
			shell:    []string{powershell, powershell6},
			command:  exeWithSpace,
			args:     []string{"arg1", "arg 2", "arg` custom` escape"},
			expected: []string{"-Command", fmt.Sprintf(`& "%s"`, exeWithSpace), "\"arg1\"", "\"arg 2\"", "arg` custom` escape"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			t.Parallel()

			var shells []string
			for _, shell := range test.shell {
				shells = append(shells, shell)
				shells = append(shells, shell+".exe")
				shells = append(shells, strings.ToUpper(shell+".exe"))
				shells = append(shells, "some/path/to/"+shell)
				shells = append(shells, "other path/to/"+shell)
			}
			for _, shell := range shells {
				t.Run(shell, func(t *testing.T) {
					t.Parallel()

					c := &Command{
						Shell:     []string{shell},
						Command:   test.command,
						Arguments: test.args,
						Environ:   []string{"PROFILE_NAME=test"},
					}
					composedArgs := getArgumentsComposer(shell)(c)
					assert.Equal(t, test.expected, composedArgs)
				})
			}
		})
	}
}

func TestVariablesRewrite(t *testing.T) {
	t.Parallel()

	mapper := func(name string) string { return fmt.Sprintf("!%s!", name) }

	tests := []struct{ in, expected string }{
		{in: "arg", expected: "arg"},
		{in: "no vars", expected: "no vars"},
		{in: "no vars: %no_unix%", expected: "no vars: %no_unix%"},
		{in: "no vars: $-no- $(123) $$ $? ${a:-1} ${a:-1}", expected: "no vars: $-no- $(123) $$ $? ${a:-1} ${a:-1}"},
		{in: "no vars: $Env:abc ${Env:abc}", expected: "no vars: $Env:abc ${Env:abc}"},
		{in: "no vars: $obj.prop $arr[0] $arr[$index]", expected: "no vars: $obj.prop $arr[0] $arr[!index!]"},
		{in: "$a", expected: "!a!"},
		{in: "${a}", expected: "!a!"},
		{in: "${_a_}", expected: "!_a_!"},
		{in: "$_a_", expected: "!_a_!"},
		{in: "$a ${b}_$c_-$d", expected: "!a! !b!_!c_!-!d!"},
		{in: "$d$x$y", expected: "!d!$x!y!"}, // known issue, won't fix
		{in: "${d}${x}${y}", expected: "!d!!x!!y!"},
	}

	frame := []string{"", "-", "]", "/", "\\", "/path/", "\\path\\", "-token", " token", " token ", " _ "}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

	cmd := NewCommand("command.go", nil)
	expected := "." + string(os.PathSeparator) + "command.go"

	for _, shell := range []string{defaultShell, bashShell, powershell, powershell6, windowsShell} {
		args := getArgumentsComposer(shell)(cmd)
		assert.Equal(t, expected, args[len(args)-1])
	}
}

func TestShellSearchPath(t *testing.T) {
	t.Parallel()

	searchList := NewCommand("echo", []string{}).getShellSearchList()
	assert.NotEmpty(t, searchList)
	for _, shell := range searchList {
		assert.NotNil(t, shellArgumentsComposerRegistry[shell])
	}
}

func TestSelectCustomShell(t *testing.T) {
	t.Parallel()

	cmd := NewCommand("sleep", []string{"1"})
	cmd.Shell = []string{mockBinary}
	shell, _, err := cmd.GetShellCommand()
	assert.Nil(t, err)
	assert.Equal(t, mockBinary, shell)

	expected := "cannot find shell: exec: \"non-existing-shell\": executable file not found in $PATH (tried non-existing-shell)"
	if platform.IsWindows() {
		expected = strings.ReplaceAll(expected, "$PATH", "%PATH%")
	}

	cmd.Shell = []string{"non-existing-shell"}
	shell, _, err = cmd.GetShellCommand()
	assert.EqualError(t, err, expected)
	assert.Empty(t, shell)
}

func TestRunShellWorkingDir(t *testing.T) {
	t.Parallel()

	command := func() string {
		if platform.IsWindows() {
			return "@echo %CD%"
		}
		return "pwd"
	}()
	temp := t.TempDir()
	buffer := new(strings.Builder)
	cmd := NewCommand(command, nil)
	cmd.Stdout = buffer
	cmd.Dir = temp
	_, _, err := cmd.Run()
	require.NoError(t, err)

	assert.Contains(t, strings.TrimSpace(buffer.String()), temp)
}

func TestRunShellEcho(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{"TestRunShellEcho"})
	cmd.Stdout = buffer
	_, _, err := cmd.Run()
	require.NoError(t, err)
	output, err := io.ReadAll(buffer)
	require.NoError(t, err)

	assert.Contains(t, string(output), "TestRunShellEcho")
}

func TestRunShellEchoWithSignalling(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := NewSignalledCommand("echo", []string{"TestRunShellEchoWithSignalling"}, c)
	cmd.Stdout = buffer
	_, _, err := cmd.Run()
	require.NoError(t, err)
	output, err := io.ReadAll(buffer)
	require.NoError(t, err)

	assert.Contains(t, string(output), "TestRunShellEchoWithSignalling")
}

// Flaky test on github linux runner but can't reproduce locally
func TestInterruptShellCommand(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("Test not running on this platform")
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)

	cmd := NewSignalledCommand(mockBinary, []string{"test", "--sleep", "3000"}, sigChan)
	cmd.Stdout = buffer

	// Will ask us to stop in 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- syscall.SIGINT
	}()
	start := time.Now()
	_, _, err := cmd.Run()
	require.Error(t, err)

	// check it ran for more than 100ms (but less than 500ms - the build agent can be very slow at times)
	duration := time.Since(start)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
	assert.Less(t, duration.Milliseconds(), int64(500))
}

func TestSetPIDCallback(t *testing.T) {
	t.Parallel()

	called := 0
	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{t.Name()})
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int32) {
		called++
	}
	_, _, err := cmd.Run()
	require.NoError(t, err)

	assert.Equal(t, 1, called)
}

func TestSetPIDCallbackWithSignalling(t *testing.T) {
	t.Parallel()

	called := 0
	buffer := &bytes.Buffer{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer signal.Reset(os.Interrupt)

	cmd := NewSignalledCommand("echo", []string{t.Name()}, c)
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int32) {
		called++
	}
	_, _, err := cmd.Run()
	require.NoError(t, err)

	assert.Equal(t, 1, called)
}

func TestSummaryDurationCommand(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	cmd := NewCommand("sleep", []string{"1"})
	if platform.IsWindows() {
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
	t.Parallel()

	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)
	cmd := NewSignalledCommand("sleep", []string{"1"}, sigChan)
	if platform.IsWindows() {
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
	t.Parallel()

	expected := "error message\n"
	if platform.IsWindows() {
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
	t.Parallel()

	expected := "error message\n"
	if platform.IsWindows() {
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
	t.Parallel()

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
	t.Parallel()

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

func TestCanAnalyseLockFailure(t *testing.T) {
	t.Parallel()

	fileName := filepath.Join(t.TempDir(), "test-restic-lock-failure")
	err := os.WriteFile(fileName, []byte(ResticLockFailureOutput), 0o600)
	require.NoError(t, err)

	cmd := NewCommand(mockBinary, []string{"test", "--stderr", fmt.Sprintf("@%s", fileName)})
	cmd.Stderr = &bytes.Buffer{}

	summary, _, err := cmd.Run()
	assert.NoError(t, err)
	assert.NotNil(t, summary.OutputAnalysis)
	assert.True(t, summary.OutputAnalysis.ContainsRemoteLockFailure())
}

func TestCanAnalyseWithCustomPattern(t *testing.T) {
	t.Parallel()

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
