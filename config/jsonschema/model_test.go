package jsonschema

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		obj              SchemaType
		expectedTypeName any
		expectedType     any
	}{
		{obj: newSchemaArray(nil), expectedTypeName: "array", expectedType: new(schemaArray)},
		{obj: newSchemaBase64String("text/plain"), expectedTypeName: "string", expectedType: new(schemaString)},
		{obj: newSchemaBool(), expectedTypeName: "boolean", expectedType: new(schemaTypeBase)},
		{obj: newSchemaNumber(false), expectedTypeName: "number", expectedType: new(schemaNumber)},
		{obj: newSchemaNumber(true), expectedTypeName: "integer", expectedType: new(schemaNumber)},
		{obj: newSchemaObject(), expectedTypeName: "object", expectedType: new(schemaObject)},
		{obj: newSchemaString(), expectedTypeName: "string", expectedType: new(schemaString)},
		{obj: newSchemaTypeList(true), expectedTypeName: nil, expectedType: new(schemaTypeList)},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.expectedTypeName), func(t *testing.T) {
			assert.IsType(t, test.expectedType, test.obj)
			base := test.obj.base()
			if base == nil {
				assert.Nil(t, test.expectedTypeName)
			} else {
				assert.Equal(t, test.expectedTypeName, base.Type)
			}
		})
	}
}

func TestArray(t *testing.T) {
	obj := newSchemaString()
	arr := newSchemaArray(obj)
	assert.Equal(t, obj, arr.Items)
}

func TestBase(t *testing.T) {
	base := &schemaTypeBase{}
	assert.Same(t, base, base.base())

	base.setDeprecated(true)
	assert.Equal(t, maybe.True().Nilable(), base.Deprecated)
	base.setDeprecated(false)
	assert.Nil(t, base.Deprecated)

	assert.Equal(t, "b", withBaseType(base, "b").Type)
	assert.Equal(t, []string{"a", "b", "c"}, withBaseType(base, "a", "b", "c").Type)
}

func TestBase64(t *testing.T) {
	b64 := newSchemaBase64String("application/octet-stream")
	assert.Equal(t, "base64", b64.ContentEncoding)
	assert.Equal(t, "application/octet-stream", b64.ContentMediaType)
}

func TestTypeSerialization(t *testing.T) {
	base := func(typeName, content string) string {
		return fmt.Sprintf(`{`+
			` "title": "t", "default": "dv", "deprecated": true, "description": "d", `+
			` "type": "%s", "enum": ["ev"], "examples": ["ex"] %s`+
			`}`, typeName, content)
	}
	tests := []struct {
		targetType SchemaType
		json       string
	}{
		{targetType: new(schemaArray), json: base("array", `, `+
			`"items": null, "minItems": 1, "maxItems": 10, "uniqueItems": true`)},
		{targetType: new(schemaNumber), json: base("number", `, `+
			`"multipleOf": 2.3, "minimum": 0.5, "maximum": 5.3, "exclusiveMinimum": 0.4, "exclusiveMaximum": 5.4`)},
		{targetType: new(schemaNumber), json: base("integer", `, `+
			`"multipleOf": 2, "minimum": 1, "maximum": 5, "exclusiveMinimum": 0, "exclusiveMaximum": 6`)},
		{targetType: new(schemaObject), json: base("object", `, "additionalProperties": true, `+
			`"patternProperties": { "regex": null }, "properties": { "name": null }, `+
			`"required": ["name"], "dependentRequired": { "name": ["otherName"] }`)},
		{targetType: new(schemaString), json: base("string", `, "minLength": 10, "maxLength": 100, `+
			`"contentEncoding": "ce", "contentMediaType": "cm", "pattern": "pt", "format": "ft"`)},
		{targetType: new(schemaTypeBase), json: base("boolean", "")},
		{targetType: new(schemaTypeList), json: `{ "anyOf": [ null ] }`},
		{targetType: new(schemaTypeList), json: `{ "oneOf": [ null ] }`},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := json.Unmarshal([]byte(test.json), test.targetType)
			assert.Nil(t, err)
			data, err := json.Marshal(test.targetType)
			assert.Nil(t, err)
			assert.Equal(t, strings.ReplaceAll(test.json, " ", ""), string(data))
		})
	}
}

func extractType[T any](t *testing.T, st any) *T {
	t.Helper()
	require.IsType(t, new(T), st)
	require.NotNil(t, st)
	return st.(*T)
}

func TestReferences(t *testing.T) {
	str := newSchemaString()
	str.Title = "first"
	nr := newSchemaNumber(true)
	nr.Title = "second"

	obj := newSchemaObject()
	obj.Properties[str.Title] = newSchemaTypeList(true, str, newSchemaArray(str))
	obj.Properties[nr.Title] = newSchemaTypeList(true, nr, newSchemaArray(nr))
	obj.Properties["third"] = newSchemaBool()

	root, err := newSchema(config.Version01, "", obj)
	assert.Nil(t, err)
	assert.Empty(t, root.Defs)

	root.createReferences(0)
	if js, e := json.MarshalIndent(root, "", "  "); e == nil {
		t.Log(string(js))
	}
	require.Len(t, root.Defs, 4)

	resolveRef := func(ref *schemaReference) SchemaType {
		refId := strings.Split(ref.Ref, "/$defs/")[1]
		return root.Defs[refId]
	}

	for name, base := range map[string]*schemaTypeBase{"first": str.base(), "second": nr.base()} {
		// Check and resolve reference to type list
		refToList := extractType[schemaReference](t, obj.Properties[name])
		typeList := extractType[schemaTypeList](t, resolveRef(refToList))
		require.Len(t, typeList.AnyOf, 2)
		assert.Len(t, typeList.OneOf, 0)

		// Test that typeList has only references to the original types
		ref := extractType[schemaReference](t, typeList.AnyOf[0])

		array := extractType[schemaArray](t, typeList.AnyOf[1])
		refInArray := extractType[schemaReference](t, array.Items)
		assert.Equal(t, ref.Ref, refInArray.Ref)

		// Test root.Defs contains the referenced type
		require.NotNil(t, resolveRef(ref))
		assert.Equal(t, base, resolveRef(ref).base())
	}
}

func TestVerify(t *testing.T) {
	testBase := func(t *testing.T, base *schemaTypeBase) {
		t.Helper()
		require.NotEmpty(t, validTypeNames)
		require.NotNil(t, base)

		for _, name := range validTypeNames {
			base.Type = name
			assert.NoError(t, base.verify())
			base.Type = []string{name, name, name}
			assert.NoError(t, base.verify())

			base.Type = []string{name, name, "invalid", name}
			assert.ErrorContains(t, base.verify(), `type name "invalid" not in (`)
		}

		base.Type = nil
		assert.ErrorContains(t, base.verify(), "expected single type or list of types, but none was specified")
		base.Type = "invalid"
		assert.ErrorContains(t, base.verify(), `type name "invalid" not in (`)
	}

	t.Run("base", func(t *testing.T) {
		testBase(t, new(schemaTypeBase))
	})

	t.Run("array", func(t *testing.T) {
		arr := newSchemaArray(nil)
		assert.ErrorContains(t, arr.verify(), `items of schemaArray is undefined`)

		arr.Items = newSchemaString()
		assert.NoError(t, arr.verify())

		testBase(t, arr.base())
	})

	t.Run("type-list", func(t *testing.T) {
		tl := newSchemaTypeList(true)
		assert.ErrorContains(t, tl.verify(), `neither anyOf nor oneOf defined`)
		tl = newSchemaTypeList(false)
		assert.ErrorContains(t, tl.verify(), `neither anyOf nor oneOf defined`)

		tl = newSchemaTypeList(true, newSchemaBool())
		assert.NoError(t, tl.verify())
		tl.OneOf = append(tl.OneOf, newSchemaBool())
		assert.ErrorContains(t, tl.verify(), `both, anyOf and oneOf defined`)
	})

	t.Run("object", func(t *testing.T) {
		obj := newSchemaObject()
		assert.NoError(t, obj.verify())

		obj.PatternProperties = map[string]SchemaType{"][": nil}
		assert.ErrorContains(t, obj.verify(), `type of "][" in patternProperties is undefined`)
		obj.PatternProperties["]["] = newSchemaString()
		assert.ErrorContains(t, obj.verify(), `patternProperties regex "][" failed to compile: error parsing regexp: missing closing ]`)
		obj.PatternProperties = map[string]SchemaType{".+": newSchemaString()}
		assert.NoError(t, obj.verify())
		obj.PatternProperties = map[string]SchemaType{"(?!negative-lookahead)": newSchemaString()}
		assert.NoError(t, obj.verify())

		obj.Properties = map[string]SchemaType{"first": nil}
		assert.ErrorContains(t, obj.verify(), `type of "first" in properties is undefined`)
		obj.Properties["first"] = newSchemaString()
		assert.NoError(t, obj.verify())

		assert.Equal(t, false, obj.AdditionalProperties)
		obj.AdditionalProperties = "-"
		assert.ErrorContains(t, obj.verify(), `additionalProperties must be nil, boolean or SchemaType`)
		obj.AdditionalProperties = newSchemaString()
		assert.NoError(t, obj.verify())

		testBase(t, obj.base())
	})

	t.Run("string", func(t *testing.T) {
		str := newSchemaString()
		assert.NoError(t, str.verify())

		str.Pattern = "]["
		assert.ErrorContains(t, str.verify(), `pattern regex "][" failed to compile: error parsing regexp: missing closing ]`)
		str.Pattern = ".+"
		assert.NoError(t, str.verify())

		require.NotEmpty(t, validFormatNames)
		str.Format = "invalid"
		assert.ErrorContains(t, str.verify(), `format "invalid" is no valid string format`)

		for _, name := range validFormatNames {
			str.Format = name
			assert.NoError(t, str.verify())
		}

		testBase(t, str.base())
	})

	t.Run("new-schema", func(t *testing.T) {
		obj := newSchemaObject()

		obj.Properties = map[string]SchemaType{"p": newSchemaArray(nil)}
		root, err := newSchema(config.Version01, "0.11", obj)
		assert.Nil(t, root)
		assert.ErrorContains(t, err, `items of schemaArray is undefined`)

		obj.Properties = map[string]SchemaType{"p": newSchemaArray(newSchemaBool())}
		root, err = newSchema(config.Version01, "0.11", obj)
		assert.NotNil(t, root)
		assert.NoError(t, err)
	})
}
