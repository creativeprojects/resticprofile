package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testGroupData struct {
	format string
	config string
}

func TestGetProfileGroupsWithNothing(t *testing.T) {
	testData := []testGroupData{
		{FormatTOML, ""},
		{FormatJSON, "{}"},
		{FormatYAML, ""},
		{FormatHCL, ""},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			groups := c.GetProfileGroups()
			assert.NotNil(t, groups)
			assert.Empty(t, groups)
		})
	}
}

func TestGetProfileGroupsWithEmpty(t *testing.T) {
	testData := []testGroupData{
		{
			FormatTOML,
			"[groups]\n",
		},
		{
			FormatJSON,
			`{ "groups": { } }`,
		},
		{
			FormatYAML,
			"\ngroups:\n",
		},
		{
			FormatHCL,
			"groups = { }",
		},
		{
			FormatTOML,
			"version = 2\n[groups]\n",
		},
		{
			FormatJSON,
			`{ "version": 2, "groups": { } }`,
		},
		{
			FormatYAML,
			"\nversion: 1\ngroups:\n",
		},
		{
			FormatHCL,
			"version = 2\ngroups = { }",
		},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(testItem.format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			groups := c.GetProfileGroups()
			assert.NotNil(t, groups)
			assert.Empty(t, groups)
		})
	}
}

func TestGetProfileGroups(t *testing.T) {
	testData := []testGroupData{
		{FormatTOML, `[groups]
first = ["backup"]
second = ["root", "dev"]
`},
		{FormatJSON, `{"groups": {"first": ["backup"], "second": ["root","dev"]}}`},
		{FormatYAML, `---
groups:
  first: "backup"
  second: ["root", "dev"]
`},
		{FormatHCL, `
"groups" = {
	"first" = ["backup"]
	second = ["root","dev"]
}
`},
		{FormatHCL, `
"groups" = {
	"first" = ["backup"]
}
groups = {
	second = ["root","dev"]
}
`},
		{FormatTOML, `version = 2
[groups]
[groups.first]
profiles = ["backup"]
[groups.second]
profiles = ["root", "dev"]
`},
		{FormatJSON, `{"version": 2, "groups": {"first": {"profiles": ["backup"]}, "second": {"profiles": ["root","dev"]}}}`},
		{FormatYAML, `---
version: 2
groups:
  first:
    profiles: "backup"
  second:
    profiles: ["root", "dev"]
`},
	}
	for _, fixture := range testData {
		t.Run(fixture.format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(fixture.config), fixture.format)
			require.NoError(t, err)

			groups := c.GetProfileGroups()
			assert.NotNil(t, groups)
			assert.Len(t, groups, 2)
			assert.Len(t, groups["first"].Profiles, 1)
			assert.Len(t, groups["second"].Profiles, 2)
		})
	}
}

func TestGetProfileGroup(t *testing.T) {
	testData := []testGroupData{
		{
			FormatTOML,
			`
[groups]
test = ["first", "second", "third"]
`,
		},
		{
			FormatJSON,
			`{ "groups": { "test": ["first", "second", "third"] } }`,
		},
		{
			FormatYAML,
			`
groups:
  test:
  - first
  - second
  - third
`,
		},
		{
			FormatHCL,
			`
groups = {
	"test" = ["first", "second", "third"]
}
`,
		},
		{
			FormatTOML,
			`
version = 2
[groups]
[groups.test]
profiles = ["first", "second", "third"]
`,
		},
		{
			FormatJSON,
			`{ "version": 2, "groups": { "test": {"profiles": ["first", "second", "third"] } } }`,
		},
		{
			FormatYAML,
			`
version: 2
groups:
  test:
    profiles:
    - first
    - second
    - third
`,
		},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(testItem.format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			assert.False(t, c.HasProfileGroup("my-group"))
			assert.True(t, c.HasProfileGroup("test"))

			group, err := c.GetProfileGroup("test")
			require.NoError(t, err)
			assert.Equal(t, &Group{
				config:   c,
				Name:     "test",
				Profiles: []string{"first", "second", "third"},
			}, group)

			_, err = c.GetProfileGroup("my-group")
			assert.Error(t, err)
		})
	}
}
