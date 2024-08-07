package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacyArgModifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		argValue              string
		argType               ArgType
		expectedLegacyArgType ArgType
	}{
		{"0 to 4", ArgConfigEscape, ArgLegacyEscape},
		{"1 to 5", ArgConfigKeepGlobQuote, ArgLegacyKeepGlobQuote},
		{"2 to 6", ArgCommandLineEscape, ArgLegacyCommandLineEscape},
		{"3 to 7", ArgConfigBackupSource, ArgLegacyConfigBackupSource},
	}

	for _, testCase := range testCases {
		t.Run(testCase.argValue, func(t *testing.T) {
			t.Parallel()

			t.Run("ON", func(t *testing.T) {
				t.Parallel()

				modifier := NewLegacyArgModifier(true)
				arg := NewArg(testCase.argValue, testCase.argType)
				newArg, changed := modifier.Arg("", &arg)
				assert.Equal(t, true, changed)
				assert.Equal(t, testCase.expectedLegacyArgType, newArg.Type())
				// run again to check if the type is not changed
				newArgAgain, changed := modifier.Arg("", newArg)
				assert.Equal(t, false, changed)
				assert.Equal(t, testCase.expectedLegacyArgType, newArgAgain.Type())
			})

			t.Run("OFF", func(t *testing.T) {
				t.Parallel()

				modifier := NewLegacyArgModifier(false)
				arg := NewArg(testCase.argValue, testCase.argType)
				newArg, changed := modifier.Arg("", &arg)
				assert.Equal(t, false, changed)
				assert.Equal(t, testCase.argType, newArg.Type())
			})
		})
	}
}
