package config

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type showStructData struct {
	input  interface{}
	output string
}

type testObject struct {
	Id      int                 `mapstructure:"id"`
	Name    string              `mapstructure:"name"`
	Person  testPerson          `mapstructure:"person"`
	Pointer *testPointer        `mapstructure:"pointer"`
	Other   string              `mapstructure:"other" show:"noshow"`
	Hidden  string              `mapstructure:""`
	Map     map[string][]string `mapstructure:",remain"`
}

type testPerson struct {
	Name       string              `mapstructure:"name"`
	IsValid    bool                `mapstructure:"valid"`
	Properties map[string][]string `mapstructure:"properties"`
}

type testStringer struct {
	Age time.Duration `mapstructure:"age"`
}

type testPointer struct {
	IsValid bool `mapstructure:"valid"`
}

type testEmbedded struct {
	EmbeddedStruct `mapstructure:",squash"`
	InlineValue    int `mapstructure:"inline"`
}

type EmbeddedStruct struct {
	Value bool `mapstructure:"value"`
}

func TestShowStruct(t *testing.T) {
	testData := []showStructData{
		{
			input:  testObject{Id: 11, Name: "test"},
			output: " id: 11\n name:  test\n",
		},
		{
			input:  testObject{Id: 11, Person: testPerson{Name: "test"}},
			output: " id:  11\n\n person:\n  name:  test\n",
		},
		{
			input:  testObject{Id: 11, Person: testPerson{Name: "test", IsValid: true}},
			output: " id:  11\n\n person:\n  name:   test\n  valid:  true\n",
		},
		{
			input:  testObject{Id: 11, Pointer: &testPointer{IsValid: true}},
			output: " id:  11\n\n pointer:\n  valid:  true\n",
		},
		{
			input: testObject{Id: 11, Person: testPerson{Properties: map[string][]string{
				"list": {"one", "two", "three"},
			}}},
			output: " id:  11\n\n person.properties:\n  list:  one\n      two\n      three\n",
		},
		{
			input:  testObject{Id: 11, Name: "test", Map: map[string][]string{"left": {"over"}}},
			output: " id: 11\n name:  test\n left:  over\n",
		},
		{
			input:  testObject{Id: 11, Name: "test", Other: "should not appear", Hidden: "should not appear either"},
			output: " id: 11\n name:  test\n",
		},
		{
			input:  testEmbedded{EmbeddedStruct{Value: true}, 1},
			output: " value:   true\n inline:  1\n",
		},
		{
			input:  &testEmbedded{EmbeddedStruct{Value: true}, 1},
			output: " value:   true\n inline:  1\n",
		},
		{
			input:  testStringer{Age: 2*time.Minute + 5*time.Second},
			output: " age:  2m5s\n",
		},
		{
			input:  &testStringer{Age: 2*time.Minute + 5*time.Second},
			output: " age:  2m5s\n",
		},
		{
			input:  testStringer{},
			output: "",
		},
	}

	for _, testItem := range testData {
		t.Run("", func(t *testing.T) {
			b := &strings.Builder{}
			err := ShowStruct(b, testItem.input, "top-level")
			assert.NoError(t, err)
			assert.Equal(t, "top-level:\n"+testItem.output, strings.ReplaceAll(b.String(), "    ", " "))
		})
	}
}

func TestUnsupportedSliceShowStruct(t *testing.T) {
	input := []showStructData{
		{
			input:  testObject{Id: 11, Name: "test"},
			output: " id: 11\n name:  test\n",
		},
	}
	b := &strings.Builder{}
	err := ShowStruct(b, input, "invalid")
	assert.Error(t, err)
}

func TestUnsupportedMapShowStruct(t *testing.T) {
	input := map[string]showStructData{
		"first": {
			input:  testObject{Id: 11, Name: "test"},
			output: " id: 11\n name:  test\n",
		},
	}
	b := &strings.Builder{}
	err := ShowStruct(b, input, "invalid")
	assert.Error(t, err)
}
