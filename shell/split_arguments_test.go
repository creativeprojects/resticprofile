package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitArguments(t *testing.T) {
	testCases := []struct {
		commandLine  string
		expectedArgs []string
	}{
		{
			commandLine:  `cmd arg1 arg2`,
			expectedArgs: []string{"cmd", "arg1", "arg2"},
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
		},
		{
			commandLine:  `cmd arg\ with\ spaces`,
			expectedArgs: []string{"cmd", "arg with spaces"},
		},
	}

	for _, testCase := range testCases {
		args := SplitArguments(testCase.commandLine)
		assert.Equal(t, testCase.expectedArgs, args)
	}
}
