package schedule

import (
	"testing"
)

func TestRawArgs(t *testing.T) {
	args := []string{"arg1", "arg2"}
	ca := NewCommandArguments(args)
	rawArgs := ca.RawArgs()
	if len(rawArgs) != len(args) {
		t.Errorf("expected %d raw arguments, got %d", len(args), len(rawArgs))
	}
	for i, arg := range args {
		if rawArgs[i] != arg {
			t.Errorf("expected raw argument %d to be %s, got %s", i, arg, rawArgs[i])
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		args     []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"arg1"}, "arg1"},
		{[]string{"arg1 with space"}, `"arg1 with space"`},
		{[]string{"arg1", "arg2"}, "arg1 arg2"},
		{[]string{"arg1", "arg with spaces"}, `arg1 "arg with spaces"`},
		{[]string{"arg1", "arg with spaces", "anotherArg"}, `arg1 "arg with spaces" anotherArg`},
	}

	for _, test := range tests {
		ca := NewCommandArguments(test.args)
		result := ca.String()
		if result != test.expected {
			t.Errorf("expected %s, got %s", test.expected, result)
		}
	}
}
