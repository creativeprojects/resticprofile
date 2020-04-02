package main

import (
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

func TestShellCommand(t *testing.T) {
	command := "/bin/restic"
	args := []string{
		`-v`,
		`--exclude-file`,
		`"excludes"`,
		`--repo`,
		`"/Volumes/RAMDisk"`,
		`backup`,
		`.`,
	}
	cmd, err := getShellCommand(command, args)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		assert.Equal(t, []string{
			`C:\Windows\system32\cmd.exe`,
			`/C`,
			`/bin/restic`,
			`-v`,
			`--exclude-file`,
			`excludes`,
			`--repo`,
			`/Volumes/RAMDisk`,
			`backup`,
			`.`,
		}, cmd.Args)
	} else {
		assert.Equal(t, []string{
			"/bin/sh",
			"-c",
			"/bin/restic -v --exclude-file \"excludes\" --repo \"/Volumes/RAMDisk\" backup .",
		}, cmd.Args)
	}
}
