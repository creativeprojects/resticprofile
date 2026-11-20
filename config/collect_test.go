package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectStruct(t *testing.T) {
	testData := []struct {
		name   string
		input  any
		expect map[string]any
	}{
		{
			name:  "simple fields",
			input: testObject{Id: 11, Name: "test"},
			expect: map[string]any{
				"id":   "11",
				"name": "test",
			},
		},
		{
			name:  "nested struct",
			input: testObject{Id: 11, Person: testPerson{Name: "test", IsValid: true}},
			expect: map[string]any{
				"id": "11",
				"person": map[string]any{
					"name":  "test",
					"valid": true,
				},
			},
		},
		{
			name:  "nested struct empty fields hidden",
			input: testObject{Id: 11, Person: testPerson{Name: "test"}},
			expect: map[string]any{
				"id": "11",
				"person": map[string]any{
					"name": "test",
				},
			},
		},
		{
			name:  "pointer struct",
			input: testObject{Id: 11, Pointer: &testPointer{IsValid: true}},
			expect: map[string]any{
				"id": "11",
				"pointer": map[string]any{
					"valid": true,
				},
			},
		},
		{
			name: "slice of structs",
			input: testObject{Id: 11, Persons: []testPerson{
				{Name: "p1", IsValid: true},
				{Name: "p2"},
			}},
			expect: map[string]any{
				"id": "11",
				"persons": []any{
					map[string]any{"name": "p1", "valid": true},
					map[string]any{"name": "p2"},
				},
			},
		},
		{
			name:  "remain map merged at top level",
			input: testObject{Id: 11, Map: map[string]any{"extra": "value"}},
			expect: map[string]any{
				"id":    "11",
				"extra": "value",
			},
		},
		{
			name:  "show tag field",
			input: testObject{Id: 11, OtherShown: "shown"},
			expect: map[string]any{
				"id":          "11",
				"other-shown": "shown",
			},
		},
		{
			name: "hidden fields not included",
			input: testObject{
				Id:         11,
				Other:      "should not appear",
				Hidden:     "should not appear",
				AlsoHidden: "should not appear",
			},
			expect: map[string]any{
				"id": "11",
			},
		},
		{
			name:  "embedded squash struct",
			input: testEmbedded{EmbeddedStruct{Value: true}, 1},
			expect: map[string]any{
				"value":  true,
				"inline": "1",
			},
		},
		{
			name:  "stringer value",
			input: testStringer{Age: 2*time.Minute + 5*time.Second},
			expect: map[string]any{
				"age": "2m5s",
			},
		},
		{
			name:   "zero stringer hidden",
			input:  testStringer{},
			expect: map[string]any{},
		},
		{
			name:   "maybe.Bool unset hidden",
			input:  testObject{IsValid: maybe.Bool{}},
			expect: map[string]any{},
		},
		{
			name:  "maybe.Bool false shown",
			input: testObject{IsValid: maybe.False()},
			expect: map[string]any{
				"valid": "false",
			},
		},
		{
			name:  "maybe.Bool true shown",
			input: testObject{IsValid: maybe.True()},
			expect: map[string]any{
				"valid": "true",
			},
		},
		{
			name:  "pointer to struct",
			input: &testEmbedded{EmbeddedStruct{Value: true}, 1},
			expect: map[string]any{
				"value":  true,
				"inline": "1",
			},
		},
		{
			name:  "allowed empty value args",
			input: testObject{Map: map[string]any{"tag": "", "keep-tag": "", "group-by": ""}},
			expect: map[string]any{
				"tag":      "",
				"keep-tag": "",
				"group-by": "",
			},
		},
		{
			name: "nested map in remain",
			input: testObject{Map: map[string]any{
				"nested": map[string]any{
					"key1": "val1",
					"key2": []string{"a", "b"},
				},
			}},
			expect: map[string]any{
				"nested": map[string]any{
					"key1": "val1",
					"key2": []any{"a", "b"},
				},
			},
		},
		{
			name: "nested map with deeper nesting",
			input: testObject{Map: map[string]any{
				"outer": map[string]any{
					"inner": map[string]any{
						"deep": "value",
					},
				},
			}},
			expect: map[string]any{
				"outer": map[string]any{
					"inner": map[string]any{
						"deep": "value",
					},
				},
			},
		},
		{
			name: "list of mixed types in remain",
			input: testObject{Map: map[string]any{
				"items": []any{
					"plain",
					map[string]any{"a": "b"},
				},
			}},
			expect: map[string]any{
				"items": []any{
					"plain",
					map[string]any{"a": "b"},
				},
			},
		},
		{
			name:  "nested struct in map",
			input: testObject{Map: map[string]any{"person": testPerson{Name: "test", IsValid: true}}},
			expect: map[string]any{
				"person": map[string]any{
					"name":  "test",
					"valid": true,
				},
			},
		},
		{
			name: "map with string properties",
			input: testObject{Id: 11, Person: testPerson{Properties: map[string][]string{
				"list": {"one", "two", "three"},
			}}},
			expect: map[string]any{
				"id": "11",
				"person": map[string]any{
					"properties": map[string]any{
						"list": []any{"one", "two", "three"},
					},
				},
			},
		},
	}

	for i, testItem := range testData {
		t.Run(fmt.Sprintf("%d_%s", i, testItem.name), func(t *testing.T) {
			result, err := CollectStruct(testItem.input)
			require.NoError(t, err)
			assert.Equal(t, testItem.expect, result)
		})
	}
}

func TestCollectStructUnsupportedType(t *testing.T) {
	_, err := CollectStruct([]string{"not", "a", "struct"})
	assert.Error(t, err)
}

func TestCollectStructNilPointer(t *testing.T) {
	var ptr *testObject
	result, err := CollectStruct(ptr)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
