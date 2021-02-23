package shell

import (
	"bytes"
	"io/ioutil"
	"os"
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
		`"/Volumes/RAMDisk"`,
		`backup`,
		`.`,
	}
	command, args, err := getShellCommand(testCommand, testArgs)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			`/C`,
			`/bin/restic`,
			`-v`,
			`--exclude-file`,
			`excludes`,
			`--repo`,
			`/Volumes/RAMDisk`,
			`backup`,
			`.`,
		}, args)
	} else {
		assert.Regexp(t, regexp.MustCompile("(/usr)?/bin/sh"), command)
		assert.Equal(t, []string{
			"-c",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup .",
		}, args)
	}
}

func TestShellCommand(t *testing.T) {
	testCommand := "/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup ."
	testArgs := []string{}

	command, args, err := getShellCommand(testCommand, testArgs)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		assert.Equal(t, `c:\windows\system32\cmd.exe`, strings.ToLower(command))
		assert.Equal(t, []string{
			"/C",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup .",
		}, args)
	} else {
		assert.Regexp(t, regexp.MustCompile("(/usr)?/bin/sh"), command)
		assert.Equal(t, []string{
			"-c",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup .",
		}, args)
	}
}

func TestRunShellEcho(t *testing.T) {
	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{"TestRunShellEcho"})
	cmd.Stdout = buffer
	_, err := cmd.Run()
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
	_, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, string(output), "TestRunShellEchoWithSignalling")
}

func TestInterruptShellCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test not running on this platform")
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)

	cmd := NewSignalledCommand("sleep", []string{"3"}, sigChan)
	cmd.Stdout = buffer

	// Will ask us to stop in 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		sigChan <- syscall.Signal(syscall.SIGINT)
	}()
	start := time.Now()
	_, err := cmd.Run()
	// GitHub Actions *sometimes* sends a different message: "signal: interrupt"
	if err != nil && err.Error() != "exit status 1" && err.Error() != "signal: interrupt" {
		t.Fatal(err)
	}

	// check it ran for more than 100ms (but less than 150ms)
	duration := time.Since(start)
	assert.GreaterOrEqual(t, duration.Milliseconds(), int64(100))
	assert.Less(t, duration.Milliseconds(), int64(150))
}

func TestSetPIDCallback(t *testing.T) {
	called := 0
	buffer := &bytes.Buffer{}
	cmd := NewCommand("echo", []string{"TestSetPIDCallback"})
	cmd.Stdout = buffer
	cmd.SetPID = func(pid int) {
		called++
	}
	_, err := cmd.Run()
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
	_, err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 1, called)
}

func TestSummaryDurationCommand(t *testing.T) {
	// if runtime.GOOS == "windows" {
	// 	t.Skip("Test not running on this platform")
	// }
	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	cmd := NewCommand("sleep", []string{"1"})
	cmd.Stdout = buffer

	start := time.Now()
	summary, err := cmd.Run()
	require.NoError(t, err)

	// make sure the command ran properly
	assert.WithinDuration(t, time.Now(), start.Add(1*time.Second), 100*time.Millisecond)
	assert.GreaterOrEqual(t, summary.Duration.Milliseconds(), int64(1000))
	assert.Less(t, summary.Duration.Milliseconds(), int64(1100))
}

func TestSummaryDurationSignalledCommand(t *testing.T) {
	// if runtime.GOOS == "windows" {
	// 	t.Skip("Test not running on this platform")
	// }
	if testing.Short() {
		t.SkipNow()
	}
	buffer := &bytes.Buffer{}

	sigChan := make(chan os.Signal, 1)
	cmd := NewSignalledCommand("sleep", []string{"1"}, sigChan)
	cmd.Stdout = buffer

	start := time.Now()
	summary, err := cmd.Run()
	require.NoError(t, err)

	// make sure the command ran properly
	assert.WithinDuration(t, time.Now(), start.Add(1*time.Second), 100*time.Millisecond)
	assert.GreaterOrEqual(t, summary.Duration.Milliseconds(), int64(1000))
	assert.Less(t, summary.Duration.Milliseconds(), int64(1100))
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
