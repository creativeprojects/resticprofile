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
		{"toml", ""},
		{"json", "{}"},
		{"yaml", ""},
		{"hcl", ""},
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
			"toml",
			"[groups]\n",
		},
		{
			"json",
			`{ "groups": { } }`,
		},
		{
			"yaml",
			"\ngroups:\n",
		},
		{
			"hcl",
			`groups = { }`,
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
	testConfig := `[groups]
first = ["backup"]
second = ["root", "dev"]
`
	c, err := Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	groups := c.GetProfileGroups()
	assert.NotNil(t, groups)
	assert.Len(t, groups, 2)
}
func TestGetProfileGroup(t *testing.T) {
	testData := []testGroupData{
		{
			"toml",
			`
[groups]
test = ["first", "second", "third"]
`,
		},
		{
			"json",
			`{ "groups": { "test": ["first", "second", "third"] } }`,
		},
		{
			"yaml",
			`
groups:
  test:
  - first
  - second
  - third
`,
		},
		{
			"hcl",
			`
groups = {
	"test" = ["first", "second", "third"]
}
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
			assert.Equal(t, []string{"first", "second", "third"}, group)

			_, err = c.GetProfileGroup("my-group")
			assert.Error(t, err)
		})
	}
}
