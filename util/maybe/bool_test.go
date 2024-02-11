package maybe_test

import (
	"strconv"
	"testing"

	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
)

func TestMaybeBool(t *testing.T) {
	fixtures := []struct {
		source             maybe.Bool
		isTrue             bool
		isStrictlyFalse    bool
		isFalseOrUndefined bool
		isUndefined        bool
		isTrueOrUndefined  bool
	}{
		{
			source:             maybe.Bool{},
			isTrue:             false,
			isStrictlyFalse:    false,
			isFalseOrUndefined: true,
			isUndefined:        true,
			isTrueOrUndefined:  true,
		},
		{
			source:             maybe.True(),
			isTrue:             true,
			isStrictlyFalse:    false,
			isFalseOrUndefined: false,
			isUndefined:        false,
			isTrueOrUndefined:  true,
		},
		{
			source:             maybe.False(),
			isTrue:             false,
			isStrictlyFalse:    true,
			isFalseOrUndefined: true,
			isUndefined:        false,
			isTrueOrUndefined:  false,
		},
	}

	for _, fixture := range fixtures {
		t.Run(toString(fixture.source), func(t *testing.T) {
			assert.Equal(t, fixture.isTrue, fixture.source.IsTrue())
			assert.Equal(t, fixture.isStrictlyFalse, fixture.source.IsStrictlyFalse())
			assert.Equal(t, fixture.isFalseOrUndefined, fixture.source.IsFalseOrUndefined())
			assert.Equal(t, fixture.isUndefined, fixture.source.IsUndefined())
			assert.Equal(t, fixture.isTrueOrUndefined, fixture.source.IsTrueOrUndefined())
		})
	}
}

func toString(value maybe.Bool) string {
	if !value.HasValue() {
		return "<nil>"
	}
	return strconv.FormatBool(value.Value())
}
