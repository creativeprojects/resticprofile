package shell

import (
	"bytes"
	"io/ioutil"
	"runtime"
	"testing"

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
	assert.Equal(t, "TestRunShellEcho\n", string(output))
}
