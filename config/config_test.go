package config

import (
	"bytes"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyAllKeysForTOML(t *testing.T) {
	testConfig := ""

	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)
	keys := configuration.AllKeys()
	assert.Empty(t, keys)
}

func TestAllKeysForTOML(t *testing.T) {
	testConfig := `
[profile1]
value = true
[profile2]
value = false
[profile3]
[profile3.backup]
[profile3.retention]
[profile4]
value = 1
[profile4.backup]
source = "/"
[profile5]
other = 2
[profile5.snapshots]
`
	expected := `profile1.value
profile2.value
profile4.backup.source
profile4.value
profile5.other`

	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)
	keys := configuration.AllKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	assert.Equal(t, expected, strings.Join(keys, "\n"))
}

func TestNoProfileGroups(t *testing.T) {
	testConfig := ""

	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	assert.Nil(t, configuration.ProfileGroups())
}

func TestEmptyProfileGroups(t *testing.T) {
	testConfig := `[groups]
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	assert.NotNil(t, configuration.ProfileGroups())
}

func TestProfileGroups(t *testing.T) {
	testConfig := `[groups]
first = ["backup"]
second = ["root", "dev"]
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	groups := configuration.ProfileGroups()
	assert.NotNil(t, groups)
	assert.Len(t, groups, 2)
}

func TestNoProfileSectionsForTOML(t *testing.T) {
	testConfig := ""

	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	assert.Nil(t, configuration.ProfileSections())
}

func TestProfileSectionsForTOML(t *testing.T) {
	testConfig := `
[profile1]
[profile2]
[profile3]
[profile3.backup]
[profile3.retention]
[profile4]
value = 1
[profile4.backup]
source = "/"
[profile5]
other = 2
[profile5.snapshots]
[global]
Initialize = true
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	profileSections := configuration.ProfileSections()
	assert.NotNil(t, profileSections)
	assert.Len(t, profileSections, 2)
}

func TestGetGlobalFromJSON(t *testing.T) {
	testConfig := `
{
  "global": {
    "default-command": "version",
    "initialize": false,
    "priority": "low"
  }
}`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "json")
	require.NoError(t, err)

	global, err := GetGlobalSection(configuration)
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}

func TestGetGlobalFromYAML(t *testing.T) {
	testConfig := `
global:
    default-command: version
    initialize: false
    priority: low
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "yaml")
	require.NoError(t, err)

	global, err := GetGlobalSection(configuration)
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}

func TestGetGlobalFromTOML(t *testing.T) {
	testConfig := `
[global]
priority = "low"
default-command = "version"
# initialize a repository if none exist at location
initialize = false
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	global, err := GetGlobalSection(configuration)
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}

func TestGetGlobalFromHCL(t *testing.T) {
	testConfig := `
"global" = {
    default-command = "version"
    initialize = false
    priority = "low"
}
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "hcl")
	require.NoError(t, err)

	global, err := GetGlobalSection(configuration)
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}

func TestGetGlobalFromSplitConfig(t *testing.T) {
	testConfig := `
"global" = {
    default-command = "version"
    initialize = true
}

"global" = {
    initialize = false
    priority = "low"
}
`
	configuration := NewConfig()
	err := configuration.Load(bytes.NewBufferString(testConfig), "hcl")
	require.NoError(t, err)

	global, err := GetGlobalSection(configuration)
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}
