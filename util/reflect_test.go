package util

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnpackValue(t *testing.T) {
	var tests = []struct {
		value    any
		expected reflect.Type
	}{
		{value: "v", expected: reflect.TypeFor[string]()},
		{value: new("v"), expected: reflect.TypeFor[string]()},
		{value: new(io.Writer(nil)), expected: nil},
		{value: new(io.Writer(&strings.Builder{})), expected: reflect.TypeFor[strings.Builder]()},
		{value: io.Writer(&strings.Builder{}), expected: reflect.TypeFor[strings.Builder]()},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			value, isNil := UnpackValue(reflect.ValueOf(test.value))
			if test.expected == nil {
				assert.True(t, isNil)
			} else {
				assert.False(t, isNil)
				assert.Equal(t, test.expected, value.Type())
			}
		})
	}
}

func TestElementType(t *testing.T) {
	var tests = []struct {
		value    any
		expected reflect.Type
	}{
		{value: "v", expected: reflect.TypeFor[string]()},
		{value: map[int]string{}, expected: reflect.TypeFor[string]()},
		{value: &map[int]string{}, expected: reflect.TypeFor[string]()},
		{value: []string{}, expected: reflect.TypeFor[string]()},
		{value: append([]string{}, "v"), expected: reflect.TypeFor[string]()},
		{value: make(chan string), expected: reflect.TypeFor[string]()},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			value := ElementType(reflect.TypeOf(test.value))
			assert.Equal(t, test.expected, value)
		})
	}
}
