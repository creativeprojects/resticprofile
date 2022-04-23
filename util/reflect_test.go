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
		{value: "v", expected: reflect.TypeOf("")},
		{value: CopyRef("v"), expected: reflect.TypeOf("")},
		{value: CopyRef(io.Writer(nil)), expected: nil},
		{value: CopyRef(io.Writer(&strings.Builder{})), expected: reflect.TypeOf(strings.Builder{})},
		{value: io.Writer(&strings.Builder{}), expected: reflect.TypeOf(strings.Builder{})},
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
		{value: "v", expected: reflect.TypeOf("")},
		{value: map[int]string{}, expected: reflect.TypeOf("")},
		{value: &map[int]string{}, expected: reflect.TypeOf("")},
		{value: []string{}, expected: reflect.TypeOf("")},
		{value: append([]string{}, "v"), expected: reflect.TypeOf("")},
		{value: make(chan string), expected: reflect.TypeOf("")},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			value := ElementType(reflect.TypeOf(test.value))
			assert.Equal(t, test.expected, value)
		})
	}
}
