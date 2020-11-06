package config

import (
	"strings"
	"testing"

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
	Map     map[string][]string `mapstructure:",remain"`
}

type testPerson struct {
	Name       string              `mapstructure:"name"`
	IsValid    bool                `mapstructure:"valid"`
	Properties map[string][]string `mapstructure:"properties"`
}

type testPointer struct {
	IsValid bool `mapstructure:"valid"`
}

func TestShowStruct(t *testing.T) {
	testData := []showStructData{
		{
			input:  testObject{Id: 11, Name: "test"},
			output: " person:\n\n id: 11\n name:  test\n\n",
		},
		{
			input:  testObject{Id: 11, Person: testPerson{Name: "test"}},
			output: " person:\n  name:  test\n\n id:  11\n\n",
		},
		{
			input:  testObject{Id: 11, Person: testPerson{Name: "test", IsValid: true}},
			output: " person:\n  name:   test\n  valid:  true\n\n id:  11\n\n",
		},
		{
			input:  testObject{Id: 11, Pointer: &testPointer{IsValid: true}},
			output: " person:\n\n pointer:\n  valid:  true\n\n id:  11\n\n",
		},
		{
			input: testObject{Id: 11, Person: testPerson{Properties: map[string][]string{
				"list": {"one", "two", "three"},
			}}},
			output: " person:\n  properties:\n   list:  one\n       two\n       three\n\n\n id:  11\n\n",
		},
		{
			input:  testObject{Id: 11, Name: "test", Map: map[string][]string{"left": {"over"}}},
			output: " person:\n\n id: 11\n name:  test\n left:  over\n\n",
		},
	}

	for _, testItem := range testData {
		b := &strings.Builder{}
		err := ShowStruct(b, testItem.input)
		assert.NoError(t, err)
		assert.Equal(t, testItem.output, strings.ReplaceAll(b.String(), "    ", " "))
	}
}
