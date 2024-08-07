package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfidentialArgModifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		arg              Arg
		expectedModified bool
		expectedValue    string
	}{
		{NewEmptyValueArg(), false, ""},
		{NewArg("value", ArgConfigEscape), false, "value"},
		{NewArg("non confidential", ArgConfigEscape, NewConfidentialArgOption(false)), false, "non confidential"},
		{NewArg("non confidential", ArgConfigEscape, NewConfidentialArgOption(true)), true, "non confidential"},
		{NewArg("local:user:password@host/path", ArgConfigEscape, NewConfidentialArgOption(true)), true, "local:user:×××@host/path"},
	}

	modifier := NewConfidentialArgModifier()

	for _, testCase := range testCases {
		newArg, modified := modifier.Arg("name", &testCase.arg)
		assert.Equal(t, testCase.expectedModified, modified)
		assert.Equal(t, testCase.expectedValue, newArg.Value())
	}
}
