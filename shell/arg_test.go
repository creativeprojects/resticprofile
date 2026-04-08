package shell

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestEmptyArgValue(t *testing.T) {
	t.Parallel()

	noValue := NewArg("", ArgConfigKeepGlobQuote)
	emptyValue := NewEmptyValueArg()

	assert.False(t, noValue.HasValue())
	assert.False(t, noValue.IsEmptyValue())
	assert.Equal(t, "", noValue.Value())

	assert.True(t, emptyValue.HasValue())
	assert.True(t, emptyValue.IsEmptyValue())
	assert.Equal(t, "", emptyValue.Value())
	if !platform.IsWindows() {
		assert.Equal(t, `""`, emptyValue.String())
	}
}

func TestArgClone(t *testing.T) {
	t.Parallel()

	args := []Arg{
		NewEmptyValueArg(),
		NewArg("value", ArgConfigEscape),
		NewArg("non confidential", ArgConfigEscape, NewConfidentialArgOption(false)),
		NewArg("confidential", ArgConfigEscape, NewConfidentialArgOption(true)),
	}

	for _, arg := range args {
		t.Run(arg.Value(), func(t *testing.T) {
			t.Parallel()

			clone := arg.Clone()

			if arg.confidentialFilter == nil || clone.confidentialFilter == nil {
				assert.Equal(t, arg, clone)
			} else {
				// assert library cannot compare functions, so we need to check manually
				assert.Equal(t, arg.Value(), clone.Value())
				assert.Equal(t, arg.Type(), clone.Type())
				assert.True(t, arg.HasConfidentialFilter())
				assert.True(t, clone.HasConfidentialFilter())
				assert.Equal(t, arg.GetConfidentialValue(), clone.GetConfidentialValue())
			}
		})
	}
}
