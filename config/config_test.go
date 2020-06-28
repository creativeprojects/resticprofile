package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGlobalFromJSON(t *testing.T) {
	testConfig := `
{
  "global": {
    "default-command": "version",
    "initialize": false,
    "priority": "low"
  }
}`
	c, err := Load(bytes.NewBufferString(testConfig), "json")
	require.NoError(t, err)

	global, err := c.GetGlobalSection()
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
	c, err := Load(bytes.NewBufferString(testConfig), "yaml")
	require.NoError(t, err)

	global, err := c.GetGlobalSection()
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
	c, err := Load(bytes.NewBufferString(testConfig), "toml")
	require.NoError(t, err)

	global, err := c.GetGlobalSection()
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
	c, err := Load(bytes.NewBufferString(testConfig), "hcl")
	require.NoError(t, err)

	global, err := c.GetGlobalSection()
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
	c, err := Load(bytes.NewBufferString(testConfig), "hcl")
	require.NoError(t, err)

	global, err := c.GetGlobalSection()
	require.NoError(t, err)
	assert.Equal(t, "version", global.DefaultCommand)
	assert.Equal(t, false, global.Initialize)
	assert.Equal(t, "low", global.Priority)
	assert.Equal(t, false, global.IONice)
}
