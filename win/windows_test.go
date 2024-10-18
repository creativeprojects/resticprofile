//go:build windows

package win

import (
	"testing"
)

func TestParseArguments(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "NoSpaces",
			args:     []string{"arg1", "arg2", "arg3"},
			expected: "arg1 arg2 arg3",
		},
		{
			name:     "WithSpaces",
			args:     []string{"arg1", "arg 2", "arg3"},
			expected: `arg1 "arg 2" arg3`,
		},
		{
			name:     "AllWithSpaces",
			args:     []string{"arg 1", "arg 2", "arg 3"},
			expected: `"arg 1" "arg 2" "arg 3"`,
		},
		{
			name:     "EmptyArgs",
			args:     []string{},
			expected: "",
		},
		{
			name:     "MixedArgs",
			args:     []string{"arg1", "arg 2", "arg3", "arg 4"},
			expected: `arg1 "arg 2" arg3 "arg 4"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArguments(tt.args)
			if result != tt.expected {
				t.Errorf("parseArguments(%v) = %v; expected %v", tt.args, result, tt.expected)
			}
		})
	}
}
