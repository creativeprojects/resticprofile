package config

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
)

type showStructData struct {
	input  interface{}
	output string
}

type testObject struct {
	Id         int                    `mapstructure:"id"`
	Name       string                 `mapstructure:"name"`
	Person     testPerson             `mapstructure:"person"`
	Persons    []testPerson           `mapstructure:"persons"`
	Pointer    *testPointer           `mapstructure:"pointer"`
	OtherShown string                 `show:"other-shown"`
	Other      string                 `mapstructure:"other" show:"noshow"`
	Hidden     string                 `mapstructure:""`
	AlsoHidden string                 `mapstructure:"use"`
	Map        map[string]interface{} `mapstructure:",remain"`
	RemainMap  map[string]interface{} `show:",remain"`
	IsValid    maybe.Bool             `mapstructure:"valid"`
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

	InlineValue int `mapstructure:"inline"`
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
			input:  testObject{Id: 11, Map: map[string]interface{}{"person": testPerson{Name: "test", IsValid: true}}},
			output: " id:  11\n\n person:\n  name:   test\n  valid:  true\n",
		},
		{
			input:  testObject{Id: 11, Persons: []testPerson{{Name: "p1", IsValid: true}, {Name: "p2", IsValid: false}}},
			output: " id:  11\n\n persons:\n  name:   p1\n  valid:  true\n  -\n  name:  p2\n",
		},
		{
			input:  testObject{Id: 11, Map: map[string]interface{}{"persons": []testPerson{{Name: "p1", IsValid: true}, {Name: "p2", IsValid: false}}}},
			output: " id:  11\n\n persons:\n  name:   p1\n  valid:  true\n  -\n  name:  p2\n",
		},
		{
			input:  testObject{Id: 11, Pointer: &testPointer{IsValid: true}},
			output: " id:  11\n\n pointer:\n  valid:  true\n",
		},
		{
			input:  testObject{Id: 11, OtherShown: "test"},
			output: " id:     11\n other-shown:  test\n",
		},
		{
			input: testObject{Id: 11, Person: testPerson{Properties: map[string][]string{
				"list": {"one", "two", "three"},
			}}},
			output: " id:  11\n\n person.properties:\n  list:  one\n      two\n      three\n",
		},
		{
			input:  testObject{Id: 11, Name: "test", Map: map[string]interface{}{"left": []string{"over"}}},
			output: " id: 11\n name:  test\n left:  over\n",
		},
		{
			input:  testObject{Id: 11, Name: "test", RemainMap: map[string]interface{}{"left": []string{"over"}}},
			output: " id: 11\n name:  test\n left:  over\n",
		},
		{
			input: testObject{Map: map[string]interface{}{
				// "use" is hidden
				constants.SectionConfigurationMixinUse: []string{"u1", "u2", "u3"},
				// List with map and list
				"list": []interface{}{
					"one",
					map[string]interface{}{
						"a":                                    "b",
						"x":                                    []string{"y1", "y2"},
						constants.SectionConfigurationMixinUse: "deep-use-is-shown",
					},
					"three",
				},
				// Map with map and list
				"a-map": map[string]interface{}{
					"mk1": []string{"my1", "my2"},
					"mk2": map[string]interface{}{"ck1": "cv", "ck2": []string{"cky1", "cky2"}},
				},
			}},
			output: " a-map:  mk1:{my1,my2}\n   mk2:{ck1:cv,ck2:{cky1,cky2}}\n list:   one\n   {a:b,use:deep-use-is-shown,x:{y1,y2}}\n   three\n",
		},
		{
			input: testObject{
				Id:         11,
				Name:       "test",
				Other:      "should not appear",
				Hidden:     "should not appear either",
				AlsoHidden: "should not appear either",
			},
			output: " id: 11\n name:  test\n",
		},
		{
			input: testObject{Map: map[string]interface{}{
				"tag":      "", // special field should show empty string
				"keep-tag": "", // special field should show empty string
				"group-by": "", // special field should show empty string
				"other":    "", // otherwise we don't show empty string
			}},
			output: " group-by:  \"\"\n keep-tag:  \"\"\n tag:    \"\"\n",
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
		{
			input:  testObject{IsValid: maybe.Bool{}},
			output: "", // display no value
		},
		{
			input:  testObject{IsValid: maybe.False()},
			output: " valid:  false\n",
		},
		{
			input:  testObject{IsValid: maybe.True()},
			output: " valid:  true\n",
		},
	}

	for i, testItem := range testData {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
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
