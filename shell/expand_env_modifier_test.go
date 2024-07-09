package shell

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestExpandEnv(t *testing.T) {
	t.Parallel()

	if platform.IsWindows() {
		t.Skip("Not running on Windows")
	}

	testCases := []struct {
		environment  []string
		value        string
		expected     string
		shouldChange bool
	}{
		{[]string{}, "", "", false},
		{[]string{}, "something", "something", false},
		{[]string{"wrong"}, "$wrong", "$wrong", false},                       // no environment variable
		{[]string{}, "$notfound", "$notfound", false},                        // value not updated at all
		{[]string{"found=true"}, "$notfound$found", "${notfound}true", true}, // value partially updated
		{[]string{"empty="}, "$empty", "", true},
		{[]string{"key=value"}, "$key", "value", true},
		{[]string{"key = value"}, "$key", "value", true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.value, func(t *testing.T) {
			t.Parallel()

			modifier := NewExpandEnvModifier(testCase.environment)
			arg := NewArg(testCase.value, ArgConfigEscape)
			newArg, changed := modifier.Arg("", &arg)
			assert.Equal(t, testCase.shouldChange, changed)
			assert.Equal(t, testCase.expected, newArg.Value())
		})
	}
}
