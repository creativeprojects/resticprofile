package shell

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestSplitArguments(t *testing.T) {
	testCases := []struct {
		commandLine  string
		expectedArgs []string
		windowsOnly  bool
		unixOnly     bool
	}{
		{
			commandLine:  `cmd arg1 arg2`,
			expectedArgs: []string{"cmd", "arg1", "arg2"},
		},
		{
			commandLine:  ``,
			expectedArgs: []string{},
		},
		{
			commandLine:  `cmd "arg with spaces" arg3`,
			expectedArgs: []string{"cmd", "arg with spaces", "arg3"},
		},
		{
			commandLine:  `cmd "arg with spaces" "another arg"`,
			expectedArgs: []string{"cmd", "arg with spaces", "another arg"},
		},
		{
			commandLine:  `cmd "arg with spaces"`,
			expectedArgs: []string{"cmd", "arg with spaces"},
		},
		{
			commandLine:  `cmd`,
			expectedArgs: []string{"cmd"},
		},
		{
			commandLine:  `"cmd file"`,
			expectedArgs: []string{"cmd file"},
		},
		{
			commandLine:  `"cmd file" arg`,
			expectedArgs: []string{"cmd file", "arg"},
		},
		{
			commandLine:  `cmd "arg \"with\" spaces"`,
			expectedArgs: []string{"cmd", "arg \"with\" spaces"},
			unixOnly:     true,
		},
		{
			commandLine:  `cmd arg\ with\ spaces`,
			expectedArgs: []string{"cmd", "arg with spaces"},
			unixOnly:     true,
		},
		{
			commandLine:  `args --with folder/file.txt`,
			expectedArgs: []string{"args", "--with", "folder/file.txt"},
		},
		{
			commandLine:  `args --with folder\file.txt`,
			expectedArgs: []string{"args", "--with", "folder\\file.txt"},
			windowsOnly:  true,
		},
	}

	for _, testCase := range testCases {
		if testCase.windowsOnly && !platform.IsWindows() {
			continue
		}
		if testCase.unixOnly && platform.IsWindows() {
			continue
		}
		t.Run(testCase.commandLine, func(t *testing.T) {
			args := SplitArguments(testCase.commandLine)
			assert.Equal(t, testCase.expectedArgs, args)
		})
	}
}
