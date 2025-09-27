//go:build !ajv_test

package jsonschema

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/config"
	"github.com/creativeprojects/resticprofile/config/mocks"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func compileSchema(t *testing.T, version config.Version) *jsonschema.Schema {
	t.Helper()

	schemaBuffer := &bytes.Buffer{}
	err := WriteJsonSchema(version, "", schemaBuffer)
	require.NoError(t, err)
	schemaJSON, err := jsonschema.UnmarshalJSON(schemaBuffer)
	require.NoError(t, err)
	compiler := jsonschema.NewCompiler()
	err = compiler.AddResource("schema.json", schemaJSON)
	require.NoError(t, err)
	schema, err := compiler.Compile("schema.json")
	require.NoError(t, err)
	return schema
}

func TestJSONSchemaCompilation(t *testing.T) {
	t.Parallel()

	for _, version := range []config.Version{config.Version01, config.Version02} {
		t.Run(fmt.Sprintf("config-%v", version), func(t *testing.T) {
			compileSchema(t, version)
		})
	}
}

func TestJSONSchemaValidation(t *testing.T) {
	t.Parallel()

	schema1 := compileSchema(t, config.Version01)
	schema2 := compileSchema(t, config.Version02)

	rewriteToJson := func(t *testing.T, filename string) string {
		t.Helper()
		v := viper.New()
		v.SetConfigFile(filename)
		if strings.HasSuffix(filename, ".conf") {
			v.SetConfigType("toml")
		}
		require.NoError(t, v.ReadInConfig())

		v.SetConfigType("json")
		filename = filepath.Join(t.TempDir(), filepath.Base(filename)+".json")
		require.NoError(t, v.WriteConfigAs(filename))
		return filename
	}

	extensionMatcher := regexp.MustCompile(`\.(conf|toml|yaml|json)$`)
	version2Matcher := regexp.MustCompile(`"version":\s*"2`)

	t.Run("examples", func(t *testing.T) {
		t.Parallel()

		exclusions := regexp.MustCompile(`[\\/](rsyslogd\.conf|utf.*\.conf|drop-in-example\.conf)$`)
		testCount := 0

		err := filepath.Walk("../../examples/", func(filename string, info fs.FileInfo, err error) error {
			if !info.IsDir() && extensionMatcher.MatchString(filename) && !exclusions.MatchString(filename) {
				content, e := os.ReadFile(filename)
				require.NoError(t, e)
				if bytes.Contains(content, []byte("{{ define")) || bytes.Contains(content, []byte("{{ template")) {
					return nil // skip test for templates
				}
				testCount++

				t.Run(path.Base(filename), func(t *testing.T) {
					if !strings.HasSuffix(filename, ".json") {
						filename = rewriteToJson(t, filename)
					}
					filename, _ = filepath.Abs(filename)

					content, e = os.ReadFile(filename)
					assert.NoError(t, e)
					schema := schema1
					if version2Matcher.Find(content) != nil {
						schema = schema2
					}

					t.Parallel()
					contentObject, err := jsonschema.UnmarshalJSON(bytes.NewReader(content))
					require.NoError(t, err)

					err = schema.Validate(contentObject)
					if err != nil {
						t.Log(string(content))
					}

					require.NoError(t, err)
				})
			}
			return err
		})

		assert.NoError(t, err)
		assert.Greater(t, testCount, 8)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		err := filepath.Walk("./", func(filename string, info fs.FileInfo, err error) error {
			if !info.IsDir() && extensionMatcher.MatchString(filename) {
				content, e := os.ReadFile(filename)
				require.NoError(t, e)

				t.Run(path.Base(filename), func(t *testing.T) {
					if !strings.HasSuffix(filename, ".json") {
						filename = rewriteToJson(t, filename)
					}
					filename, _ = filepath.Abs(filename)

					content, e = os.ReadFile(filename)
					assert.NoError(t, e)
					schema := schema1
					if version2Matcher.Find(content) != nil {
						schema = schema2
					}

					t.Parallel()
					contentObject, err := jsonschema.UnmarshalJSON(bytes.NewReader(content))
					require.NoError(t, err)

					err = schema.Validate(contentObject)
					if err == nil {
						t.Log(string(content))
					}

					require.Error(t, err)
				})
			}
			return err
		})

		assert.NoError(t, err)
	})
}

func TestValueTypeConversion(t *testing.T) {
	t.Parallel()

	boolType := newSchemaBool()
	intType := newSchemaNumber(true)
	numType := newSchemaNumber(false)
	strType := newSchemaString()

	tests := []struct {
		targetType    SchemaType
		value         string
		valueType     any
		compat, isDef bool
	}{
		{targetType: boolType, value: `true`, valueType: true, compat: true},
		{targetType: boolType, value: `false`, valueType: true, compat: true, isDef: true},
		{targetType: boolType, value: `"true"`, valueType: "", compat: false},
		{targetType: boolType, value: `anything`, valueType: "", compat: false},
		{targetType: boolType, value: `1`, valueType: int64(0), compat: false},
		{targetType: intType, value: `true`, valueType: true, compat: false},
		{targetType: intType, value: `"true"`, valueType: "", compat: false},
		{targetType: intType, value: `0`, valueType: int64(0), compat: true, isDef: true},
		{targetType: intType, value: `1`, valueType: int64(0), compat: true},
		{targetType: intType, value: `1.2`, valueType: float64(0), compat: false},
		{targetType: numType, value: `true`, valueType: true, compat: false},
		{targetType: numType, value: `"true"`, valueType: "", compat: false},
		{targetType: numType, value: `anything`, valueType: "", compat: false},
		{targetType: numType, value: `1`, valueType: int64(0), compat: true},
		{targetType: numType, value: `1.2`, valueType: float64(0), compat: true},
		{targetType: numType, value: `0.0`, valueType: float64(0), compat: true, isDef: true},
		{targetType: numType, value: `0`, valueType: int64(0), compat: true, isDef: true},
		{targetType: strType, value: `true`, valueType: true, compat: true},
		{targetType: strType, value: `"true"`, valueType: "", compat: true},
		{targetType: strType, value: `anything`, valueType: "", compat: true},
		{targetType: strType, value: `1`, valueType: int64(0), compat: true},
		{targetType: strType, value: `1.2`, valueType: float64(0), compat: true},
		{targetType: strType, value: ``, valueType: "", compat: true, isDef: true},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d_%s", i, test.value), func(t *testing.T) {
			t.Parallel()

			value := convertToType(test.value)
			assert.IsType(t, test.valueType, value, "type %q", test.value)
			assert.Equal(t, test.compat, isCompatibleValue(test.targetType, value), "value compat %q", test.value)
			assert.Equal(t, test.isDef, isDefaultValueForType(value), "value isDef %q", test.value)
		})
	}
}

var propertyInfoDefaults = map[string]any{
	"CanBeNil":            false,
	"CanBeBool":           false,
	"CanBeNumeric":        false,
	"CanBeString":         false,
	"CanBePropertySet":    false,
	"IsDeprecated":        false,
	"IsSingle":            false,
	"IsSinglePropertySet": false,
	"IsMultiType":         false,
	"IsOption":            false,
	"IsRequired":          false,
	"Name":                "",
	"Description":         "",
	"DefaultValue":        []string{""},
	"EnumValues":          nil,
	"ExampleValues":       nil,
}

var propertySetDefaults = map[string]any{
	"TypeName":          "",
	"IsClosed":          false,
	"IsAllOptions":      false,
	"Properties":        nil,
	"OtherPropertyInfo": nil,
	"Name":              "",
	"Description":       "",
}

func setupMock(t *testing.T, m *mock.Mock, defs map[string]any) {
	t.Helper()
	for method, value := range defs {
		if !m.IsMethodCallable(t, method) {
			m.On(method).Return(value).Maybe()
		}
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()

	newInfo := func(option *restic.Option, deprecated bool) *mocks.PropertyInfo {
		info := new(mocks.PropertyInfo)
		info.EXPECT().Description().Return("property-description")
		info.EXPECT().IsOption().Return(option != nil)
		if option != nil {
			info.EXPECT().Option().Return(*option)
		}
		info.EXPECT().IsDeprecated().Return(deprecated)
		return info
	}

	assertDescription := func(t *testing.T, expected string, info *mocks.PropertyInfo) {
		t.Helper()
		assert.Equal(t, expected, getDescription(info))
		info.AssertExpectations(t)
	}

	t.Run("simple", func(t *testing.T) {
		t.Parallel()

		info := newInfo(nil, false)
		assertDescription(t, "property-description", info)
	})

	t.Run("deprecated", func(t *testing.T) {
		t.Parallel()

		info := newInfo(nil, true)
		assertDescription(t, "property-description [deprecated]", info)
	})

	t.Run("removed-option", func(t *testing.T) {
		t.Parallel()

		info := newInfo(&restic.Option{RemovedInVersion: "1.24"}, true)
		assertDescription(t, "property-description [deprecated, removed in 1.24]", info)
	})

	t.Run("new-option", func(t *testing.T) {
		t.Parallel()

		info := newInfo(&restic.Option{FromVersion: "1.6"}, false)
		assertDescription(t, "property-description [restic >= 1.6]", info)
	})

	t.Run("legacy-option", func(t *testing.T) {
		t.Parallel()

		info := newInfo(&restic.Option{FromVersion: "1.2", RemovedInVersion: "1.6"}, true)
		assertDescription(t, "property-description [deprecated, removed in 1.6, restic >= 1.2]", info)
	})
}

func boolProperty(info *mocks.PropertyInfo) *mocks.PropertyInfo {
	info.EXPECT().CanBeBool().Return(true)
	return info
}

func numberProperty(info *mocks.PropertyInfo, mustInteger bool, constraint config.NumericRange) *mocks.PropertyInfo {
	info.EXPECT().CanBeNumeric().Return(true)
	info.EXPECT().MustBeInteger().Return(mustInteger)
	info.EXPECT().NumericRange().Return(constraint)
	return info
}

func stringProperty(info *mocks.PropertyInfo, format, pattern string) *mocks.PropertyInfo {
	info.EXPECT().CanBeString().Return(true)
	info.EXPECT().Format().Return(format)
	info.EXPECT().ValidationPattern().Return(pattern)
	return info
}

func objectProperty(info *mocks.PropertyInfo, set config.NamedPropertySet) *mocks.PropertyInfo {
	info.EXPECT().CanBePropertySet().Return(true)
	info.EXPECT().PropertySet().Return(set)
	return info
}

func TestSchemaForPropertySet(t *testing.T) {
	t.Parallel()

	newMock := func(config func(m *mocks.NamedPropertySet)) *mocks.NamedPropertySet {
		nps := new(mocks.NamedPropertySet)
		config(nps)
		setupMock(t, &nps.Mock, propertySetDefaults)
		return nps
	}

	t.Run("AdditionalProperties", func(t *testing.T) {
		t.Parallel()

		s := schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) { m.EXPECT().IsClosed().Return(false) }))
		assert.Equal(t, true, s.AdditionalProperties)
		s = schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) { m.EXPECT().IsClosed().Return(true) }))
		assert.Equal(t, false, s.AdditionalProperties)
	})

	t.Run("TypedAdditionalProperty", func(t *testing.T) {
		t.Parallel()

		pi := new(mocks.PropertyInfo)
		stringProperty(pi, "", "")
		pi.EXPECT().IsSingle().Return(true)
		pi.EXPECT().IsSinglePropertySet().Return(true)
		setupMock(t, &pi.Mock, propertyInfoDefaults)

		s := schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) {
			m.EXPECT().IsClosed().Return(false)
			m.EXPECT().OtherPropertyInfo().Return(pi)
		}))

		assert.Equal(t, false, s.AdditionalProperties)
		assert.Equal(t, newSchemaString(), s.PatternProperties[matchAll])
	})

	t.Run("Title", func(t *testing.T) {
		t.Parallel()

		s := schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) { m.EXPECT().Name().Return("t123") }))
		assert.Equal(t, "t123", s.Title)
	})

	t.Run("Description", func(t *testing.T) {
		t.Parallel()

		s := schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) { m.EXPECT().Description().Return("d123") }))
		assert.Equal(t, "d123", s.Description)
	})

	t.Run("Properties", func(t *testing.T) {
		t.Parallel()

		ps := new(mocks.NamedPropertySet)
		setupMock(t, &ps.Mock, propertySetDefaults)

		singleProperty := func(required bool) *mocks.PropertyInfo {
			pi := new(mocks.PropertyInfo)
			pi.EXPECT().IsSingle().Return(true)
			pi.EXPECT().IsSinglePropertySet().Return(true)
			pi.EXPECT().IsRequired().Return(required)
			return pi
		}

		props := map[string]*mocks.PropertyInfo{
			"single-str":    stringProperty(singleProperty(false), "date", ".+"),
			"single-num":    numberProperty(singleProperty(true), false, config.NumericRange{}),
			"multiple-str":  stringProperty(new(mocks.PropertyInfo), "", ""),
			"single-nested": objectProperty(singleProperty(true), ps),
			"nil":           nil,
		}

		schema := schemaForPropertySet(newMock(func(m *mocks.NamedPropertySet) {
			m.EXPECT().Properties().Return(slices.Collect(maps.Keys(props)))
			for name, info := range props {
				if info != nil {
					setupMock(t, &info.Mock, propertyInfoDefaults)
					m.EXPECT().PropertyInfo(name).Return(info)
				} else {
					m.EXPECT().PropertyInfo(name).Return(nil)
				}
			}
		}))

		assert.Len(t, schema.Properties, len(props)-1)

		t.Run("single-str", func(t *testing.T) {
			t.Parallel()

			require.IsType(t, &schemaString{}, schema.Properties["single-str"])
			if sp, ok := schema.Properties["single-str"].(*schemaString); ok {
				assert.Equal(t, stringFormat("date"), sp.Format)
				assert.Equal(t, ".+", sp.Pattern)
			}
			assert.NotContains(t, schema.Required, "single-str")
		})

		t.Run("single-num", func(t *testing.T) {
			t.Parallel()

			assert.IsType(t, &schemaNumber{}, schema.Properties["single-num"])
			assert.Contains(t, schema.Required, "single-num")
		})

		t.Run("single-nested", func(t *testing.T) {
			t.Parallel()

			assert.IsType(t, &schemaObject{}, schema.Properties["single-nested"])
			assert.Contains(t, schema.Required, "single-nested")
		})

		t.Run("multiple-str", func(t *testing.T) {
			t.Parallel()

			require.IsType(t, &schemaTypeList{}, schema.Properties["multiple-str"])
			if tl, ok := schema.Properties["multiple-str"].(*schemaTypeList); ok {
				assert.IsType(t, &schemaString{}, tl.AnyOf[0])
				require.IsType(t, &schemaArray{}, tl.AnyOf[1])
				assert.Same(t, tl.AnyOf[0], tl.AnyOf[1].(*schemaArray).Items)
			}
		})
	})
}

func TestTypesFromPropertyInfo(t *testing.T) {
	t.Parallel()

	nr := config.NumericRange{}
	ps := new(mocks.NamedPropertySet)
	setupMock(t, &ps.Mock, propertySetDefaults)

	np := func(nr config.NumericRange) *mocks.PropertyInfo {
		return numberProperty(new(mocks.PropertyInfo), false, nr)
	}

	sp := func(format, pattern string) *mocks.PropertyInfo {
		return stringProperty(new(mocks.PropertyInfo), format, pattern)
	}

	tests := []struct {
		info   *mocks.PropertyInfo
		types  []any
		nested int
		check  func(t *testing.T, types []SchemaType)
	}{
		// single-type
		{info: boolProperty(new(mocks.PropertyInfo)), types: []any{"boolean"}},
		{info: numberProperty(new(mocks.PropertyInfo), false, nr), types: []any{"number"}},
		{info: numberProperty(new(mocks.PropertyInfo), true, nr), types: []any{"integer"}},
		{info: sp("", ""), types: []any{"string"}},
		{info: objectProperty(new(mocks.PropertyInfo), ps), types: []any{"object"}, nested: 1},

		// multi-type
		{info: numberProperty(boolProperty(new(mocks.PropertyInfo)), false, nr), types: []any{"boolean", "number"}},
		{info: numberProperty(boolProperty(new(mocks.PropertyInfo)), true, nr), types: []any{"boolean", "integer"}},
		{info: stringProperty(numberProperty(new(mocks.PropertyInfo), true, nr), "", ""), types: []any{"integer", "string"}},
		{info: stringProperty(objectProperty(new(mocks.PropertyInfo), ps), "", ""), types: []any{"string", "object"}, nested: 2},

		// number range
		{info: np(config.NumericRange{From: util.CopyRef(1.0)}), types: []any{"number"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			n := types[0].(*schemaNumber)
			assert.Equal(t, util.CopyRef(1.0), n.Minimum)
			assert.Nil(t, n.Maximum)
			assert.Nil(t, n.ExclusiveMinimum)
			assert.Nil(t, n.ExclusiveMaximum)
		}},
		{info: np(config.NumericRange{To: util.CopyRef(1.0)}), types: []any{"number"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			n := types[0].(*schemaNumber)
			assert.Nil(t, n.Minimum)
			assert.Equal(t, util.CopyRef(1.0), n.Maximum)
			assert.Nil(t, n.ExclusiveMinimum)
			assert.Nil(t, n.ExclusiveMaximum)
		}},
		{info: np(config.NumericRange{
			From:          util.CopyRef(0.1),
			To:            util.CopyRef(1.0),
			FromExclusive: true,
			ToExclusive:   true,
		}), types: []any{"number"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			n := types[0].(*schemaNumber)
			assert.Equal(t, util.CopyRef(0.1), n.ExclusiveMinimum)
			assert.Equal(t, util.CopyRef(1.0), n.ExclusiveMaximum)
			assert.Nil(t, n.Minimum)
			assert.Nil(t, n.Maximum)
		}},

		// string
		{info: sp("some-format", ""), types: []any{"string"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			s := types[0].(*schemaString)
			assert.Equal(t, stringFormat("some-format"), s.Format) // validation is not performed here
		}},
		{info: sp("duration", ""), types: []any{"string"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			s := types[0].(*schemaString)
			assert.Equal(t, stringFormat(""), s.Format)
			assert.Equal(t, durationPattern, s.Pattern)
		}},
		{info: sp("duration", "custom-pattern"), types: []any{"string"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			s := types[0].(*schemaString)
			assert.Equal(t, stringFormat(""), s.Format)
			assert.Equal(t, "custom-pattern", s.Pattern)
		}},
		{info: sp("", "]some-pattern["), types: []any{"string"}, check: func(t *testing.T, types []SchemaType) {
			t.Helper()
			s := types[0].(*schemaString)
			assert.Equal(t, "]some-pattern[", s.Pattern) // validation is not performed here
		}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()

			setupMock(t, &test.info.Mock, propertyInfoDefaults)

			types, index := typesFromPropertyInfo(test.info)
			names := collect.From(types, func(t SchemaType) any { return t.base().Type })

			require.Equal(t, test.types, names)
			require.Equal(t, test.nested-1, index)

			if test.check != nil {
				test.check(t, types)
			}
			test.info.AssertExpectations(t)
		})
	}
}

func TestDurationPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		duration string
		expected time.Duration
	}{
		{"", 0},
		{"5", 0},
		{"2y", 0},
		{"2d", 0},
		{"2h", 2 * time.Hour},
		{"2m", 2 * time.Minute},
		{"2s", 2 * time.Second},
		{"1m1s", time.Second + time.Minute},
		{"1s1m", time.Second + time.Minute},
		{"1s1m1000", 0},
		{"120s5h3000ms4m", 120*time.Second + 5*time.Hour + 3000*time.Millisecond + 4*time.Minute},
	}
	pattern := regexp.MustCompile(durationPattern)
	for _, test := range tests {
		t.Run(test.duration, func(t *testing.T) {
			t.Parallel()

			duration, err := time.ParseDuration(test.duration)
			if test.expected == 0 {
				assert.False(t, pattern.MatchString(test.duration))
				assert.Error(t, err)
			} else {
				assert.True(t, pattern.MatchString(test.duration))
				assert.NoError(t, err)
				assert.Equal(t, test.expected, duration)
			}
		})
	}
}

func TestConfigureBasicInfo(t *testing.T) {
	t.Parallel()

	newType := func() SchemaType {
		return newSchemaTypeList(true, newSchemaString(), newSchemaBool(), newSchemaNumber(true))
	}
	newMock := func(method string, rv any) config.PropertyInfo {
		info := new(mocks.PropertyInfo)
		info.On(method).Return(rv)
		setupMock(t, &info.Mock, propertyInfoDefaults)
		return info
	}
	each := func(start SchemaType, fn func(item SchemaType)) {
		walkTypes(start, func(item SchemaType) SchemaType {
			if item.base() != nil {
				fn(item)
			}
			return item
		})
	}

	t.Run("Default", func(t *testing.T) {
		t.Parallel()

		schemaType := newType()
		configureBasicInfo(schemaType, nil, newMock("DefaultValue", []string{"1", "abc", "true", "3", "false", "0", ""}))
		count := 0
		each(schemaType, func(item SchemaType) {
			count++
			switch base := item.base(); base.Type {
			case "string":
				assert.Equal(t, []any{int64(1), "abc", true, int64(3)}, base.Default)
			case "integer":
				assert.Equal(t, []any{int64(1), int64(3)}, base.Default)
			case "boolean":
				assert.Equal(t, true, base.Default)
			default:
				count--
			}
		})
		assert.Equal(t, 3, count)
	})

	t.Run("Enum", func(t *testing.T) {
		t.Parallel()

		schemaType := newType()
		configureBasicInfo(schemaType, nil, newMock("EnumValues", []string{"1", "abc", "true", "3"}))
		each(schemaType, func(item SchemaType) {
			assert.Equal(t, []any{int64(1), "abc", true, int64(3)}, item.base().Enum)
		})
	})

	t.Run("Examples", func(t *testing.T) {
		t.Parallel()

		schemaType := newType()
		configureBasicInfo(schemaType, nil, newMock("ExampleValues", []string{"1", "abc", "true", "3", "false", "0", ""}))
		count := 0
		each(schemaType, func(item SchemaType) {
			count++
			switch base := item.base(); base.Type {
			case "string":
				assert.Equal(t, []any{int64(1), "abc", true, int64(3), false, int64(0), ""}, base.Examples)
			case "integer":
				assert.Equal(t, []any{int64(1), int64(3), int64(0)}, base.Examples)
			case "boolean":
				assert.Equal(t, []any{true, false}, base.Examples)
			default:
				count--
			}
		})
		assert.Equal(t, 3, count)
	})

	t.Run("Title", func(t *testing.T) {
		t.Parallel()

		schemaType := newType()
		configureBasicInfo(schemaType, nil, newMock("Name", "n1"))
		each(schemaType, func(item SchemaType) {
			assert.Equal(t, "n1", item.base().Title)
		})
	})

	t.Run("ArrayTitle", func(t *testing.T) {
		t.Parallel()

		schemaType := newSchemaArray(nil)
		configureBasicInfo(schemaType, nil, newMock("Name", "n1"))
		each(schemaType, func(item SchemaType) {
			assert.Equal(t, "n1...", item.base().Title)
		})
	})

	t.Run("Deprecated", func(t *testing.T) {
		t.Parallel()

		schemaType := newType()
		configureBasicInfo(schemaType, nil, newMock("IsDeprecated", true))
		each(schemaType, func(item SchemaType) {
			assert.Equal(t, maybe.True().Nilable(), item.base().Deprecated)
		})
	})

	t.Run("Nested", func(t *testing.T) {
		t.Parallel()

		nested := newSchemaObject()
		schemaType := newSchemaTypeList(true, newSchemaString(), nested)
		configureBasicInfo(schemaType, nested, newMock("Name", "n1"))
		assert.Equal(t, "n1", schemaType.AnyOf[0].base().Title)
		assert.Equal(t, "", schemaType.AnyOf[1].base().Title) // nested type is not modified
	})
}

func TestSchemaForConfigVersion(t *testing.T) {
	t.Parallel()

	t.Run("v1", func(t *testing.T) {
		t.Parallel()

		s := schemaForConfigVersion(config.Version01).(*schemaString)
		assert.Equal(t, uint64(0), s.MinLength)
		assert.Equal(t, util.CopyRef(uint64(1)), s.MaxLength)
		assert.Equal(t, version1Pattern, s.Pattern)
		assert.Equal(t, "1", s.Default)
		assert.True(t, regexp.MustCompile(s.Pattern).MatchString(s.Default.(string)))
	})
	t.Run("v2", func(t *testing.T) {
		t.Parallel()

		s := schemaForConfigVersion(config.Version02).(*schemaString)
		assert.Equal(t, uint64(1), s.MinLength)
		assert.Nil(t, s.MaxLength)
		assert.Equal(t, version2Pattern, s.Pattern)
		assert.Equal(t, "2", s.Default)
		assert.True(t, regexp.MustCompile(s.Pattern).MatchString(s.Default.(string)))
	})
}
