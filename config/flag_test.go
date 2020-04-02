package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointerValueShouldPanic(t *testing.T) {
	concrete := "test"
	value := &concrete
	assert.PanicsWithError(t, "Unexpected type ptr", func() {
		stringifyValueOf(value)
	})
}

func TestNilValueFlag(t *testing.T) {
	var value interface{} = nil
	argValue, hasValue := stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{}, argValue)
}

func TestTrueBooleanFlag(t *testing.T) {
	value := true
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{}, argValue)
}

func TestFalseBooleanFlag(t *testing.T) {
	value := false
	argValue, hasValue := stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{}, argValue)
}

func TestZeroIntFlag(t *testing.T) {
	value := 0
	argValue, hasValue := stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{"0"}, argValue)
}

func TestPositiveIntFlag(t *testing.T) {
	value := 1234567890
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"1234567890"}, argValue)
}

func TestNegativeIntFlag(t *testing.T) {
	value := -1234567890
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"-1234567890"}, argValue)
}

func TestEmptyStringFlag(t *testing.T) {
	value := ""
	argValue, hasValue := stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{""}, argValue)
}

func TestStringFlag(t *testing.T) {
	value := "test"
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"test"}, argValue)
}

func TestPositiveFloatFlag(t *testing.T) {
	value := 3.2
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Contains(t, argValue[0], "3.2")
}

func TestArrayOfArrayOfValueShouldPanic(t *testing.T) {
	value := [][]string{{"one", "two"}, {"three", "four"}}
	assert.PanicsWithError(t, "Array of array of values are not supported", func() {
		stringifyValueOf(value)
	})
}

func TestArrayOfStringValueFlag(t *testing.T) {
	value := [2]string{"one", "two"}
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"one", "two"}, argValue)
}

func TestEmptySliceOfStringValueFlag(t *testing.T) {
	value := []string{}
	argValue, hasValue := stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{}, argValue)
}

func TestSliceOfStringValueFlag(t *testing.T) {
	value := []string{"one", "two"}
	argValue, hasValue := stringifyValueOf(value)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"one", "two"}, argValue)
}

type testStruct struct {
	someBool1   bool   `argument:"some-bool-1"`
	someBool2   bool   `argument:"some-bool-2"`
	someString1 string `argument:"some-string-1"`
	someString2 string `argument:"some-string-2"`
	someInt1    int    `argument:"some-int-1"`
	someInt2    int    `argument:"some-int-2"`
	notIncluded bool
}

func TestConvertStructToFlag(t *testing.T) {
	testElement := &testStruct{
		someBool1:   true,
		someBool2:   false,
		someInt1:    0,
		someInt2:    10,
		someString1: "",
		someString2: "test",
		notIncluded: true,
	}
	flags := convertStructToFlags(testElement)
	assert.NotNil(t, flags)
}
