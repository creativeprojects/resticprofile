package maybe_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/creativeprojects/resticprofile/util"
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
		t.Run(fixture.source.String(), func(t *testing.T) {
			assert.Equal(t, fixture.isTrue, fixture.source.IsTrue())
			assert.Equal(t, fixture.isStrictlyFalse, fixture.source.IsStrictlyFalse())
			assert.Equal(t, fixture.isFalseOrUndefined, fixture.source.IsFalseOrUndefined())
			assert.Equal(t, fixture.isUndefined, fixture.source.IsUndefined())
			assert.Equal(t, fixture.isTrueOrUndefined, fixture.source.IsTrueOrUndefined())
		})
	}
}

func TestBoolCreators(t *testing.T) {
	assert.Equal(t, maybe.Bool{}, maybe.UnsetBool())
	assert.Equal(t, maybe.True(), maybe.SetBool(true))
	assert.Equal(t, maybe.False(), maybe.SetBool(false))

	assert.Equal(t, maybe.UnsetBool(), maybe.BoolFromNilable(nil))
	assert.Equal(t, maybe.True(), maybe.BoolFromNilable(util.CopyRef(true)))
	assert.Equal(t, maybe.False(), maybe.BoolFromNilable(util.CopyRef(false)))

	assert.Nil(t, maybe.UnsetBool().Nilable())
	assert.Equal(t, util.CopyRef(true), maybe.True().Nilable())
	assert.Equal(t, util.CopyRef(false), maybe.False().Nilable())
}

func TestBoolDecoder(t *testing.T) {
	fixtures := []struct {
		from     reflect.Type
		to       reflect.Type
		source   any
		expected any
	}{
		{
			from:     reflect.TypeOf(struct{}{}),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   struct{}{},
			expected: struct{}{}, // same value returned as the "from" type in unexpected
		},
		{
			from:     reflect.TypeOf(true),
			to:       reflect.TypeOf(""),
			source:   false,
			expected: false, // same value returned as the "to" type in unexpected
		},
		{
			from:     reflect.TypeOf(""),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   "",
			expected: "", // same value returned as the original value in unexpected
		},
		{
			from:     reflect.TypeOf(""),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   "true",
			expected: maybe.True(),
		},
		{
			from:     reflect.TypeOf(""),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   "t",
			expected: maybe.True(),
		},
		{
			from:     reflect.TypeOf(""),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   "f",
			expected: maybe.False(),
		},
		{
			from:     reflect.TypeOf(true),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   true,
			expected: maybe.True(),
		},
		{
			from:     reflect.TypeOf(true),
			to:       reflect.TypeOf(maybe.Bool{}),
			source:   false,
			expected: maybe.False(),
		},
	}
	for i, fixture := range fixtures {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			decoder := maybe.BoolDecoder()
			decoded, err := decoder(fixture.from, fixture.to, fixture.source)
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, decoded)
		})
	}
}

func TestBoolJSON(t *testing.T) {
	fixtures := []struct {
		source   maybe.Bool
		expected string
	}{
		{
			source:   maybe.UnsetBool(),
			expected: "null",
		},
		{
			source:   maybe.True(),
			expected: "true",
		},
		{
			source:   maybe.False(),
			expected: "false",
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.source.String(), func(t *testing.T) {
			// encode value into JSON
			encoded, err := fixture.source.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, fixture.expected, string(encoded))

			// decode value from JSON
			decodedValue := maybe.Bool{}
			err = decodedValue.UnmarshalJSON(encoded)
			assert.NoError(t, err)
			assert.Equal(t, fixture.source, decodedValue)
		})
	}
}

func TestBoolString(t *testing.T) {
	fixtures := []struct {
		source   maybe.Bool
		expected string
	}{
		{source: maybe.UnsetBool(), expected: ""},
		{source: maybe.True(), expected: "true"},
		{source: maybe.False(), expected: "false"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.source.String(), func(t *testing.T) {
			assert.Equal(t, fixture.expected, fixture.source.String())
		})
	}
}
