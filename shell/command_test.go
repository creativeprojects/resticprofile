package shell

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, `C:\Windows\system32\cmd.exe`, command)
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
		assert.Equal(t, "/bin/sh", command)
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
		assert.Equal(t, `C:\Windows\system32\cmd.exe`, command)
		assert.Equal(t, []string{
			"/C",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup .",
		}, args)
	} else {
		assert.Equal(t, "/bin/sh", command)
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
	err := cmd.Run()
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

	cmd := NewSignalledCommand("echo", []string{"TestRunShellEcho"}, c)
	cmd.Stdout = buffer
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, string(output), "TestRunShellEcho")
}

func TestInterruptShellCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Test not running under Windows")
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
	err := cmd.Run()
	if err != nil && err.Error() != "signal: interrupt" {
		t.Fatal(err)
	}

	assert.WithinDuration(t, time.Now(), start, 1*time.Second)
}
