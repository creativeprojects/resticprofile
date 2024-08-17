package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPropertyNamesAreSorted(t *testing.T) {
	tests := [][]string{
		(&profileInfo{sections: map[string]SectionInfo{
			"c": nil,
			"b": nil,
			"a": nil,
		}}).Sections(),
		(&propertySet{properties: map[string]PropertyInfo{
			"1c": nil,
			"1b": nil,
			"2a": nil,
		}}).Properties(),
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			assert.Len(t, test, 3)
			assert.IsIncreasing(t, test)
		})
	}
}

func TestPropertySetIsAllOptions(t *testing.T) {
	set := &propertySet{properties: map[string]PropertyInfo{
		"a": newResticPropertyInfo("a", restic.Option{}),
		"c": newResticPropertyInfo("c", restic.Option{}),
	}}
	assert.True(t, set.IsAllOptions())
	set.properties["c"] = &propertyInfo{}
	assert.False(t, set.IsAllOptions())
}

func TestResticPropertyInfo(t *testing.T) {
	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "np", newResticPropertyInfo("np", restic.Option{Name: "nr"}).Name())
	})

	t.Run("Description", func(t *testing.T) {
		assert.Equal(t, "rdesc", newResticPropertyInfo("", restic.Option{Description: "rdesc"}).Description())
	})

	t.Run("Deprecated", func(t *testing.T) {
		assert.True(t, newResticPropertyInfo("", restic.Option{RemovedInVersion: "-"}).IsDeprecated())
		assert.False(t, newResticPropertyInfo("", restic.Option{RemovedInVersion: ""}).IsDeprecated())
	})

	t.Run("IsSingle", func(t *testing.T) {
		assert.True(t, newResticPropertyInfo("", restic.Option{Once: true}).IsSingle())
		assert.False(t, newResticPropertyInfo("", restic.Option{Once: false}).IsSingle())
	})

	t.Run("DetectType", func(t *testing.T) {
		tests := []struct {
			def string
			fn  func(info PropertyInfo) bool
		}{
			{def: "true", fn: PropertyInfo.CanBeBool},
			{def: "false", fn: PropertyInfo.CanBeBool},

			{def: "0", fn: PropertyInfo.CanBeNumeric},
			{def: "0", fn: PropertyInfo.MustBeInteger},
			{def: "-1.0", fn: PropertyInfo.CanBeNumeric},
			{def: ".0", fn: PropertyInfo.CanBeNumeric},
			{def: "0.", fn: collect.Not(PropertyInfo.CanBeNumeric)},
			{def: ".", fn: collect.Not(PropertyInfo.CanBeNumeric)},
			{def: "1.0", fn: PropertyInfo.CanBeNumeric},
			{def: "1.0", fn: collect.Not(PropertyInfo.MustBeInteger)},

			{def: "\"1.0\"", fn: collect.Not(PropertyInfo.CanBeNumeric)},
			{def: "\"1.0\"", fn: collect.Not(PropertyInfo.CanBeBool)},
			{def: "\"1.0\"", fn: PropertyInfo.CanBeString},

			{def: "abc", fn: PropertyInfo.CanBeString},
		}
		for i, test := range tests {
			t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
				info := newResticPropertyInfo("", restic.Option{Default: test.def})
				assert.True(t, test.fn(info))
				assert.False(t, info.IsMultiType())
			})
		}
	})
}

func TestJoinByEmpty(t *testing.T) {
	tests := []struct {
		from, to string
	}{
		{from: "", to: ""},
		{from: ";", to: "|"},
		{from: ";;", to: ";"},
		{from: "s;;", to: "s;"},
		{from: ";;e", to: ";e"},
		{from: "a;b;;;c", to: "a|b;|c"},
		{from: "a;;;b;c", to: "a;|b|c"},
		{from: ";;a;;;b;c;;", to: ";a;|b|c;"},
		{from: "s;;e", to: "s;e"},
		{from: "s;;;e", to: "s;|e"},
		{from: "s;;;;e", to: "s;||e"},
		{from: "s;;;;;e", to: "s;|||e"},
		{from: "s;;;;;;e", to: "s;||||e"},
	}
	for _, test := range tests {
		t.Run(test.from, func(t *testing.T) {
			assert.Equal(t, test.to, strings.Join(joinEmpty(strings.Split(test.from, ";")), "|"))
		})
	}
}

func TestPropertySetFromType(t *testing.T) {
	type nestedData struct {
		//nolint:unused
		i1 string `mapstructure:"nested-string"`
	}

	var testData struct {
		hidden string
		an1    any       `mapstructure:"simple-any"`
		an     []any     `mapstructure:"simple-any-array"`
		b1     bool      `mapstructure:"simple-bool"`
		b2     []bool    `mapstructure:"simple-bool-array"`
		i0     *int      `mapstructure:"simple-integer-pointer"`
		i1     int       `mapstructure:"simple-integer"`
		i2     []int64   `mapstructure:"simple-integer-array"`
		f1     float64   `mapstructure:"simple-float"`
		f2     []float32 `mapstructure:"simple-float-array"`
		s1     string    `mapstructure:"simple-string"`
		s2     []string  `mapstructure:"simple-string-array"`

		d1 string `mapstructure:"simple-deprecated" deprecated:""`

		desc string `mapstructure:"description" description:"desc2"`
		def  string `mapstructure:"default" default:"dv1;dvesc;;;dv2"`
		ex   string `mapstructure:"examples" examples:"e1;e2;esc;;"`
		en   string `mapstructure:"enum" enum:";;vesc;v1;v2;v3"`
		fmt  string `mapstructure:"format" format:"uri"`
		fmti string `mapstructure:"format-invalid" format:"invalid"`
		ptn  string `mapstructure:"pattern" pattern:"[0-9]+"`

		r1 int `mapstructure:"closed-range" range:"[-100 : 200]"`
		r2 int `mapstructure:"lopen-range" range:"[ :200]"`
		r3 int `mapstructure:"ropen-range" range:"[-100: ]"`
		r4 int `mapstructure:"lex-range" range:"|-100:0]"`
		r5 int `mapstructure:"rex-range" range:"[0:200|"`

		m1 map[string]string     `mapstructure:"string-map"`
		m2 map[string]any        `mapstructure:"any-map"`
		m3 map[string]nestedData `mapstructure:"nested-map"`

		nested1 nestedData   `mapstructure:"nested"`
		nested2 []nestedData `mapstructure:"nested-array"`
	}

	set := propertySetFromType(reflect.TypeOf(testData))

	propertyInfo := func(t *testing.T, propertyName string) (info PropertyInfo) {
		require.Contains(t, set.Properties(), propertyName)
		info = set.PropertyInfo(propertyName)
		require.NotNil(t, info)
		return
	}

	t.Run("hidden", func(t *testing.T) {
		assert.NotContains(t, set.Properties(), "hidden")
	})

	t.Run("simple", func(t *testing.T) {
		type assertFunc func(info PropertyInfo) bool
		fn := func(args ...assertFunc) []assertFunc { return args }

		anyType := fn(PropertyInfo.IsMultiType, PropertyInfo.IsAnyType)
		oneType := fn(collect.Not(PropertyInfo.IsMultiType), collect.Not(PropertyInfo.IsAnyType))
		intType := fn(PropertyInfo.CanBeNumeric, PropertyInfo.MustBeInteger)

		defExpects := append(
			oneType,
			collect.Not(PropertyInfo.CanBeNil),
			collect.Not(PropertyInfo.IsDeprecated),
			collect.Not(PropertyInfo.IsRequired),
		)

		tests := []struct {
			property            string
			expects, defExpects []assertFunc
		}{
			{property: "simple-any", expects: fn(PropertyInfo.IsSingle), defExpects: anyType},
			{property: "simple-any-array", expects: fn(collect.Not(PropertyInfo.IsSingle)), defExpects: anyType},

			{property: "simple-bool", expects: fn(PropertyInfo.CanBeBool, PropertyInfo.IsSingle)},
			{property: "simple-bool-array", expects: fn(PropertyInfo.CanBeBool, collect.Not(PropertyInfo.IsSingle))},

			{property: "simple-integer", expects: fn(PropertyInfo.IsSingle), defExpects: append(defExpects, intType...)},
			{property: "simple-integer-array", expects: fn(collect.Not(PropertyInfo.IsSingle)), defExpects: append(defExpects, intType...)},
			{property: "simple-integer-pointer", expects: fn(PropertyInfo.CanBeNil, PropertyInfo.IsSingle), defExpects: append(oneType, intType...)},

			{property: "simple-float", expects: fn(PropertyInfo.CanBeNumeric, collect.Not(PropertyInfo.MustBeInteger), PropertyInfo.IsSingle)},
			{property: "simple-float-array", expects: fn(PropertyInfo.CanBeNumeric, collect.Not(PropertyInfo.MustBeInteger), collect.Not(PropertyInfo.IsSingle))},

			{property: "simple-string", expects: fn(PropertyInfo.CanBeString, PropertyInfo.IsSingle)},
			{property: "simple-string-array", expects: fn(PropertyInfo.CanBeString, collect.Not(PropertyInfo.IsSingle))},

			{property: "simple-deprecated", expects: fn(PropertyInfo.IsDeprecated), defExpects: oneType},

			{property: "string-map", expects: fn(PropertyInfo.CanBePropertySet, PropertyInfo.CanBeNil), defExpects: oneType},
			{property: "any-map", expects: fn(PropertyInfo.CanBePropertySet, PropertyInfo.CanBeNil), defExpects: oneType},
			{property: "nested-map", expects: fn(PropertyInfo.CanBePropertySet, PropertyInfo.CanBeNil), defExpects: oneType},

			{property: "nested", expects: fn(PropertyInfo.CanBePropertySet, collect.Not(PropertyInfo.CanBeNil)), defExpects: oneType},
		}

		for _, test := range tests {
			t.Run(test.property, func(t *testing.T) {
				info := propertyInfo(t, test.property)
				for i, expect := range test.expects {
					assert.True(t, expect(info), "expects #%d", i)
				}
				if test.defExpects == nil {
					test.defExpects = defExpects
				}
				for i, expect := range test.defExpects {
					assert.True(t, expect(info), "def-expects #%d", i)
				}
			})
		}
	})

	t.Run("description", func(t *testing.T) {
		assert.Equal(t, "desc2", propertyInfo(t, "description").Description())
	})

	t.Run("default", func(t *testing.T) {
		assert.Equal(t, []string{"dv1", "dvesc;", "dv2"}, propertyInfo(t, "default").DefaultValue())
		assert.Equal(t, []string{"false"}, propertyInfo(t, "simple-bool").DefaultValue())
		assert.Equal(t, []string{"0"}, propertyInfo(t, "simple-integer").DefaultValue())
		assert.Empty(t, propertyInfo(t, "simple-integer-pointer").DefaultValue())
	})

	t.Run("examples", func(t *testing.T) {
		assert.Equal(t, []string{"e1", "e2", "esc;"}, propertyInfo(t, "examples").ExampleValues())
	})

	t.Run("enum", func(t *testing.T) {
		assert.Equal(t, []string{";vesc", "v1", "v2", "v3"}, propertyInfo(t, "enum").EnumValues())
	})

	t.Run("format", func(t *testing.T) {
		assert.Equal(t, "uri", propertyInfo(t, "format").Format())
		assert.Empty(t, propertyInfo(t, "format-invalid").Format())
	})

	t.Run("format", func(t *testing.T) {
		assert.Equal(t, "[0-9]+", propertyInfo(t, "pattern").ValidationPattern())
	})

	t.Run("range", func(t *testing.T) {
		ref := util.CopyRef[float64]
		tests := []struct {
			property string
			expects  NumericRange
		}{
			{property: "closed-range", expects: NumericRange{From: ref(-100), To: ref(200)}},
			{property: "lopen-range", expects: NumericRange{FromExclusive: true, To: ref(200)}},
			{property: "ropen-range", expects: NumericRange{From: ref(-100), ToExclusive: true}},
			{property: "lex-range", expects: NumericRange{From: ref(-100), FromExclusive: true, To: ref(0)}},
			{property: "rex-range", expects: NumericRange{From: ref(0), To: ref(200), ToExclusive: true}},
		}

		for _, test := range tests {
			t.Run(test.property, func(t *testing.T) {
				assert.Equal(t, test.expects, propertyInfo(t, test.property).NumericRange())
			})
		}
	})

	t.Run("string-map", func(t *testing.T) {
		ps := propertyInfo(t, "string-map").PropertySet()
		require.NotNil(t, ps)

		assert.Empty(t, ps.TypeName())
		assert.Empty(t, ps.Properties())
		assert.False(t, ps.IsClosed())

		require.NotNil(t, ps.OtherPropertyInfo())
		assert.True(t, ps.OtherPropertyInfo().CanBeString())
		assert.False(t, ps.OtherPropertyInfo().IsMultiType())
	})

	t.Run("any-map", func(t *testing.T) {
		ps := propertyInfo(t, "any-map").PropertySet()
		require.NotNil(t, ps)

		assert.False(t, ps.IsClosed())
		assert.Nil(t, ps.OtherPropertyInfo())
	})

	t.Run("nested-map", func(t *testing.T) {
		ps := propertyInfo(t, "nested-map").PropertySet()
		require.NotNil(t, ps)

		assert.Empty(t, ps.TypeName())
		assert.Empty(t, ps.Properties())
		assert.False(t, ps.IsClosed())

		require.NotNil(t, ps.OtherPropertyInfo())
		assert.False(t, ps.OtherPropertyInfo().IsMultiType())

		nps := ps.OtherPropertyInfo().PropertySet()
		require.NotNil(t, nps)

		assert.Equal(t, "nestedData", nps.TypeName())
		assert.True(t, nps.IsClosed())
		assert.Equal(t, []string{"nested-string"}, nps.Properties())
	})

	t.Run("nested", func(t *testing.T) {
		info := propertyInfo(t, "nested")
		assert.True(t, info.IsSingle())

		ps := info.PropertySet()
		require.NotNil(t, ps)

		assert.Equal(t, "nestedData", ps.TypeName())
		assert.True(t, ps.IsClosed())
		assert.Equal(t, []string{"nested-string"}, ps.Properties())
	})

	t.Run("nested-array", func(t *testing.T) {
		info := propertyInfo(t, "nested-array")
		assert.False(t, info.IsSingle())

		ps := info.PropertySet()
		require.NotNil(t, ps)

		assert.Equal(t, "nestedData", ps.TypeName())
	})
}

func TestNewProfileInfo(t *testing.T) {
	t.Run("all-sections", func(t *testing.T) {
		info := NewProfileInfo(false)
		assert.Contains(t, info.Properties(), "insecure-tls")

		for name, _ := range NewProfile(nil, "").AllSections() {
			si := info.SectionInfo(name)
			require.NotNil(t, si, "section: %s", name)
			assert.True(t, si.IsCommandSection(), "section: %s", name)
		}
	})

	t.Run("with-defaults", func(t *testing.T) {
		info := NewProfileInfo(false)
		assert.NotContains(t, info.SectionInfo(constants.CommandBackup).Properties(), "insecure-tls")
		info = NewProfileInfo(true)
		assert.Contains(t, info.SectionInfo(constants.CommandBackup).Properties(), "insecure-tls")
	})

	t.Run("restic-arguments", func(t *testing.T) {
		info := NewProfileInfo(false)
		pi := info.PropertyInfo("repository")
		require.NotNil(t, pi)
		require.True(t, pi.IsOption())
		require.Equal(t, "repo", pi.Option().Name)
	})

	t.Run("restic-command", func(t *testing.T) {
		info := NewProfileInfo(false)
		tests := map[string]string{
			constants.SectionConfigurationRetention: "forget",
			constants.CommandBackup:                 constants.CommandBackup,
			constants.CommandCopy:                   constants.CommandCopy,
			constants.CommandLs:                     constants.CommandLs,
		}
		for section, commandName := range tests {
			t.Run(commandName, func(t *testing.T) {
				si := info.SectionInfo(section)
				require.NotNil(t, si)
				require.True(t, si.IsCommandSection())
				require.Equal(t, commandName, si.Command().GetName())
			})
		}
	})
}
