package bools

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBools(t *testing.T) {
	fixtures := []struct {
		source             *bool
		isTrue             bool
		isStrictlyFalse    bool
		isFalseOrUndefined bool
		isUndefined        bool
		isTrueOrUndefined  bool
	}{
		{
			source:             nil,
			isTrue:             false,
			isStrictlyFalse:    false,
			isFalseOrUndefined: true,
			isUndefined:        true,
			isTrueOrUndefined:  true,
		},
		{
			source:             boolPointer(true),
			isTrue:             true,
			isStrictlyFalse:    false,
			isFalseOrUndefined: false,
			isUndefined:        false,
			isTrueOrUndefined:  true,
		},
		{
			source:             boolPointer(false),
			isTrue:             false,
			isStrictlyFalse:    true,
			isFalseOrUndefined: true,
			isUndefined:        false,
			isTrueOrUndefined:  false,
		},
	}

	for _, fixture := range fixtures {
		t.Run(toString(fixture.source), func(t *testing.T) {
			assert.Equal(t, fixture.isTrue, IsTrue(fixture.source))
			assert.Equal(t, fixture.isStrictlyFalse, IsStrictlyFalse(fixture.source))
			assert.Equal(t, fixture.isFalseOrUndefined, IsFalseOrUndefined(fixture.source))
			assert.Equal(t, fixture.isUndefined, IsUndefined(fixture.source))
			assert.Equal(t, fixture.isTrueOrUndefined, IsTrueOrUndefined(fixture.source))
		})
	}
}

func boolPointer(value bool) *bool {
	output := &value
	return output
}

func toString(value *bool) string {
	if value == nil {
		return "<nil>"
	}
	return strconv.FormatBool(*value)
}
