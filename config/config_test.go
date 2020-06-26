package config

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoProfileKeys(t *testing.T) {
	testConfig := `
`
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	profiles := ProfileKeys()
	assert.Nil(t, profiles)
}

func TestProfileKeys(t *testing.T) {
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
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	profiles := ProfileKeys()
	assert.Len(t, profiles, 2)
	assert.Contains(t, profiles, "profile4")
	assert.Contains(t, profiles, "profile5")
}

func TestNoProfileGroups(t *testing.T) {
	testConfig := `
`
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	groups := ProfileGroups()
	assert.Nil(t, groups)
}

func TestEmptyProfileGroups(t *testing.T) {
	testConfig := `[groups]
`
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	groups := ProfileGroups()
	assert.NotNil(t, groups)
}

func TestProfileGroups(t *testing.T) {
	testConfig := `[groups]
first = ["backup"]
second = ["root", "dev"]
`
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	groups := ProfileGroups()
	assert.NotNil(t, groups)
	assert.Len(t, groups, 2)
}

func TestNoProfileSections(t *testing.T) {
	testConfig := `
`
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	profileSections := ProfileSections()
	assert.Nil(t, profileSections)
}

func TestProfileSections(t *testing.T) {
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
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	if err != nil {
		t.Fatal(err)
	}

	profileSections := ProfileSections()
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
	viper.Reset()
	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	require.NoError(t, err)

	global, err := GetGlobalSection()
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
	viper.Reset()
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	require.NoError(t, err)

	global, err := GetGlobalSection()
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
	viper.Reset()
	viper.SetConfigType("toml")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	require.NoError(t, err)

	global, err := GetGlobalSection()
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
	viper.Reset()
	viper.SetConfigType("hcl")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	require.NoError(t, err)

	global, err := GetGlobalSection()
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
	viper.Reset()
	viper.SetConfigType("hcl")
	err := viper.ReadConfig(bytes.NewBufferString(testConfig))
	require.NoError(t, err)

	global, err := GetGlobalSection()
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}
