package config

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/assert"
)

func TestPointerValueShouldReturnErrorMessage(t *testing.T) {
	concrete := "test"
	value := &concrete
	argValue, _ := stringifyValueOf(value)
	assert.Equal(t, []string{"ERROR: unexpected type ptr"}, argValue)
}

func TestNilValueFlag(t *testing.T) {
	var value interface{}
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
	assert.PanicsWithError(t, "array of array of values are not supported", func() {
		stringifyValueOf(value)
	})
}

func TestArrayOfArrayOfValueWithAny(t *testing.T) {
	value := [][]string{{"one", "two"}, {"three", "four"}}
	argValue, hasValue := stringifyAnyValueOf(value, false)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"{one,two}", "{three,four}"}, argValue)
}

func TestMap(t *testing.T) {
	value := map[string]interface{}{
		"k1": "b",
		"k2": []string{"c", "d"},
		"k3": map[string]interface{}{"x": 10, "y": false, "j": true},
	}
	argValue, hasValue := stringifyAnyValueOf(value, false)
	assert.True(t, hasValue)
	assert.Equal(t, []string{"k1:b", "k2:{c,d}", "k3:{j:true,x:10}"}, argValue)

	argValue, hasValue = stringifyValueOf(value)
	assert.False(t, hasValue)
	assert.Equal(t, []string{"ERROR: unexpected type map"}, argValue)
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
	someBoolTrue  bool   `argument:"some-bool-true"`
	someBoolFalse bool   `argument:"some-bool-false"`
	emptyString   string `argument:"empty-string"`
	globString    string `argument:"glob-string"`
	noGlobString  string `argument:"no-glob-string" argument-type:"no-glob"`
	emptyInt      int    `argument:"empty-int"`
	someInt       int    `argument:"some-int"`
	uInt          uint   `mapstructure:"unsigned-int" argument:"u-int"`
	notIncluded   bool
}

func TestConvertStructToArgs(t *testing.T) {
	testElement := testStruct{
		someBoolTrue:  true,
		someBoolFalse: false,
		emptyInt:      0,
		someInt:       10,
		uInt:          20,
		emptyString:   "",
		globString:    "test",
		noGlobString:  "test",
		notIncluded:   true,
	}
	args := shell.NewArgs()
	addArgsFromStruct(args, testElement)
	assert.NotNil(t, args)
	assert.Equal(t, map[string][]string{
		"some-bool-true": {},
		"some-int":       {"10"},
		"u-int":          {"20"},
		"glob-string":    {"test"},
		"no-glob-string": {"test"},
	}, args.ToMap())
}

func TestConvertPointerStructToArgs(t *testing.T) {
	testElement := &testStruct{
		someBoolTrue:  true,
		someBoolFalse: false,
		emptyInt:      0,
		someInt:       10,
		uInt:          20,
		emptyString:   "",
		globString:    "test",
		noGlobString:  "test",
		notIncluded:   true,
	}
	args := shell.NewArgs()
	addArgsFromStruct(args, testElement)
	assert.NotNil(t, args)
	assert.Equal(t, map[string][]string{
		"some-bool-true": {},
		"some-int":       {"10"},
		"u-int":          {"20"},
		"glob-string":    {"test"},
		"no-glob-string": {"test"},
	}, args.ToMap())
}

func TestGetArgAliasesFromStruct(t *testing.T) {
	testElement := testStruct{}
	for i, element := range []any{testElement, &testElement} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			aliases := argAliasesFromStruct(element)
			assert.NotNil(t, aliases)
			assert.Equal(t, map[string]string{"unsigned-int": "u-int"}, aliases)
		})
	}
}

func BenchmarkFormatInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value int64 = 123456
		str := strconv.FormatInt(value, 10)
		if str == "" {
			b.Fail()
		}
	}
}

func BenchmarkFormatUint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value uint64 = 123456
		str := strconv.FormatUint(value, 10)
		if str == "" {
			b.Fail()
		}
	}
}

func BenchmarkFormatFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value float64 = 123456
		str := strconv.FormatFloat(value, 'f', -1, 64)
		if str == "" {
			b.Fail()
		}
	}
}

func BenchmarkSprintfInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value int64 = 123456
		str := fmt.Sprintf("%d", value)
		if str == "" {
			b.Fail()
		}
	}
}

func BenchmarkSprintfFloat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var value float64 = 123456
		str := fmt.Sprintf("%f", value)
		if str == "" {
			b.Fail()
		}
	}
}
