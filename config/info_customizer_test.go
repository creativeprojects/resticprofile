package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
)

func customizeProperty(sectionName string, info PropertyInfo) PropertyInfo {
	props := map[string]PropertyInfo{info.Name(): info}
	customizeProperties(sectionName, props)
	return props[info.Name()]
}

func TestResticPropertyDescriptionFilter(t *testing.T) {
	tests := []struct {
		original, expected string
	}{
		{
			original: `be verbose (specify multiple times or a level using --verbose=n, max level/times is 4)`,
			expected: `be verbose (true for level 1 or a number for increased verbosity, max level is 4)`,
		},
		{
			original: `snapshot id to search in (can be given multiple times)`,
			expected: `snapshot id to search in`,
		},
		{
			original: `set extended option (key=value, can be specified multiple times)`,
			expected: `set extended option (key=value)`,
		},
		{
			original: `add tags for the new snapshot in the format tag[,tag,...] (can be specified multiple times).`,
			expected: `add tags for the new snapshot in the format tag[,tag,...].`,
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			info := customizeProperty("",
				newResticPropertyInfo("any", restic.Option{Description: test.original}))
			assert.Equal(t, test.expected, info.Description())
		})
	}
}

func TestGlobalDefaultCommandProperty(t *testing.T) {
	info := NewGlobalInfo().PropertyInfo("default-command")
	require.NotNil(t, info)
	assert.ElementsMatch(t, restic.CommandNamesForVersion(restic.AnyVersion), info.ExampleValues())
}

func TestResticVerboseProperty(t *testing.T) {
	orig := newResticPropertyInfo("verbose", restic.Option{})
	assert.True(t, orig.CanBeString())
	assert.False(t, orig.CanBeBool())
	assert.False(t, orig.CanBeNumeric())
	assert.False(t, orig.MustBeInteger())

	customizeProperty("", orig)
	assert.True(t, orig.CanBeString())
	assert.True(t, orig.CanBeBool())
	assert.True(t, orig.CanBeNumeric())
	assert.True(t, orig.MustBeInteger())
}

func TestHostTagPathProperty(t *testing.T) {
	examples := []string{"true", "false", `"{{property}}"`}
	note := `Boolean true is replaced with the {{property}}s from section "backup".`
	hostNote := `Boolean true is replaced with the hostname of the system.`
	backupNote := `Boolean true is unsupported in section "backup".`
	retentionHostNote := `Boolean true is replaced with the hostname that applies in section "backup".`
	defaultSuffix := ` Defaults to true in "{{section}}".`
	defaultSuffixV2 := ` Defaults to true for config version 2 in "{{section}}".`

	backup := constants.CommandBackup
	retention := constants.SectionConfigurationRetention

	tests := []struct {
		section, property, note, format string
		examples                        []string
	}{
		{section: "any", property: constants.ParameterHost, note: hostNote, format: "hostname"},
		{section: "any", property: constants.ParameterPath},
		{section: "any", property: constants.ParameterTag},

		{section: retention, property: constants.ParameterHost, note: retentionHostNote + defaultSuffixV2, format: "hostname"},
		{section: retention, property: constants.ParameterPath, note: note + defaultSuffix},
		{section: retention, property: constants.ParameterTag, note: note + defaultSuffixV2},

		{section: backup, property: constants.ParameterHost, note: hostNote + defaultSuffixV2, format: "hostname"},
		{section: backup, property: constants.ParameterPath, note: backupNote, examples: []string{"false", `"{{property}}"`}},
		{section: backup, property: constants.ParameterTag, note: backupNote, examples: []string{"false", `"{{property}}"`}},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			propertyReplacer := func(s string) string {
				s = strings.ReplaceAll(s, "{{property}}", test.property)
				s = strings.ReplaceAll(s, "{{section}}", test.section)
				return s
			}
			if test.examples == nil {
				test.examples = examples
			}
			if len(test.note) == 0 {
				test.note = note
			}

			info := customizeProperty(test.section, newResticPropertyInfo(test.property, restic.Option{}))

			assert.Equal(t, propertyReplacer(".\n"+test.note), info.Description())
			assert.Equal(t, collect.From(test.examples, propertyReplacer), info.ExampleValues())
			assert.Equal(t, test.format, info.Format())
			assert.True(t, info.CanBeString())
			assert.True(t, info.CanBeBool())
			assert.False(t, info.CanBeNumeric())
			assert.False(t, info.IsSingle())
		})
	}
}

func TestConfidentialProperty(t *testing.T) {
	var testType = struct {
		Simple ConfidentialValue            `mapstructure:"simple"`
		List   []ConfidentialValue          `mapstructure:"list"`
		Mapped map[string]ConfidentialValue `mapstructure:"mapped"`
	}{}

	set := propertySetFromType(reflect.TypeOf(testType))

	assert.ElementsMatch(t, []string{"simple", "list", "mapped"}, set.Properties())
	for _, name := range set.Properties() {
		t.Run(name+"/before", func(t *testing.T) {
			info := set.PropertyInfo(name)
			require.True(t, info.CanBePropertySet())
			if name == "mapped" {
				require.NotNil(t, info.PropertySet().OtherPropertyInfo())
				require.NotNil(t, info.PropertySet().OtherPropertyInfo().PropertySet())
				assert.Equal(t, "ConfidentialValue", info.PropertySet().OtherPropertyInfo().PropertySet().TypeName())
			} else {
				assert.Equal(t, "ConfidentialValue", info.PropertySet().TypeName())
			}
			assert.False(t, info.CanBeString())
			assert.False(t, info.IsMultiType())
		})
	}

	customizeProperties("any", set.properties)

	for _, name := range set.Properties() {
		t.Run(name, func(t *testing.T) {
			info := set.PropertyInfo(name)
			if name == "mapped" {
				require.True(t, info.CanBePropertySet())
				nested := info.PropertySet()
				assert.Nil(t, nested.OtherPropertyInfo())
				assert.Empty(t, nested.TypeName())
				assert.False(t, nested.IsClosed())
				assert.Empty(t, nested.Properties())
			} else {
				assert.True(t, info.CanBeString())
				assert.False(t, info.CanBePropertySet())
				assert.False(t, info.IsMultiType())
			}
		})
	}
}

func TestMaybeBoolProperty(t *testing.T) {
	var testType = struct {
		Simple maybe.Bool `mapstructure:"simple"`
	}{}

	set := propertySetFromType(reflect.TypeOf(testType))

	assert.ElementsMatch(t, []string{"simple"}, set.Properties())
	for _, name := range set.Properties() {
		t.Run(name+"/before", func(t *testing.T) {
			info := set.PropertyInfo(name)
			require.True(t, info.CanBePropertySet())
			assert.Equal(t, "Bool", info.PropertySet().TypeName())
			assert.False(t, info.CanBeBool())
			assert.False(t, info.IsMultiType())
		})
	}

	customizeProperties("any", set.properties)

	for _, name := range set.Properties() {
		t.Run(name, func(t *testing.T) {
			info := set.PropertyInfo(name)
			assert.True(t, info.CanBeNil())
			assert.True(t, info.CanBeBool())
			assert.False(t, info.CanBePropertySet())
			assert.False(t, info.IsMultiType())
		})
	}
}

func TestMaybeDurationProperty(t *testing.T) {
	var testType = struct {
		Simple maybe.Duration `mapstructure:"simple"`
	}{}

	set := propertySetFromType(reflect.TypeOf(testType))

	assert.ElementsMatch(t, []string{"simple"}, set.Properties())
	for _, name := range set.Properties() {
		t.Run(name+"/before", func(t *testing.T) {
			info := set.PropertyInfo(name)
			require.True(t, info.CanBePropertySet())
			assert.Equal(t, "Duration", info.PropertySet().TypeName())
			assert.False(t, info.CanBeNumeric())
			assert.False(t, info.IsMultiType())
		})
	}

	customizeProperties("any", set.properties)

	for _, name := range set.Properties() {
		t.Run(name, func(t *testing.T) {
			info := set.PropertyInfo(name)
			assert.True(t, info.CanBeNil())
			assert.True(t, info.CanBeNumeric())
			assert.True(t, info.CanBeString())
			assert.True(t, info.IsMultiType())
			assert.False(t, info.CanBeBool())
			assert.False(t, info.CanBePropertySet())
			assert.Equal(t, "duration", info.Format())
		})
	}
}

func TestScheduleProperty(t *testing.T) {
	var testType = struct {
		Schedule any `mapstructure:"schedule"`
	}{}

	set := propertySetFromType(reflect.TypeOf(testType))

	assert.ElementsMatch(t, []string{"schedule"}, set.Properties())
	for _, name := range set.Properties() {
		t.Run(name+"/before", func(t *testing.T) {
			info := set.PropertyInfo(name)
			require.True(t, info.IsAnyType())
		})
	}

	customizeProperties("any", set.properties)

	for _, name := range set.Properties() {
		t.Run(name, func(t *testing.T) {
			info := set.PropertyInfo(name)
			assert.True(t, info.CanBeNil())
			assert.True(t, info.IsMultiType())
			assert.True(t, info.CanBeString())
			assert.False(t, info.CanBeNumeric())
			assert.False(t, info.CanBeBool())

			require.True(t, info.CanBePropertySet())
			assert.Equal(t, NewScheduleConfigInfo().Name(), info.PropertySet().Name())

			assert.False(t, info.IsSingle(), "multiple strings")
			assert.True(t, info.IsSinglePropertySet(), "just one nested type")
		})
	}
}

func TestDeprecatedSection(t *testing.T) {
	var testType = struct {
		ScheduleBaseSection `mapstructure:",squash" deprecated:"true"`
	}{}

	set := propertySetFromType(reflect.TypeOf(testType))
	require.False(t, set.PropertyInfo("schedule").IsDeprecated())

	customizeProperties("any", set.properties)
	require.True(t, set.PropertyInfo("schedule").IsDeprecated())
}

func TestHelpIsExcluded(t *testing.T) {
	assert.True(t, isExcluded("*", "help"))
	assert.False(t, isExcluded("*", "any-other"))
	assert.False(t, isExcluded("named-section", "help"))
}
