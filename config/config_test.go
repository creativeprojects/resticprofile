package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTemplate struct {
	format string
	config string
}

func TestGetGlobal(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[global]
priority = "low"
default-command = "version"
# initialize a repository if none exist at location
initialize = false
`},
		{"json", `
{
  "global": {
    "default-command": "version",
    "initialize": false,
    "priority": "low"
  }
}`},
		{"yaml", `---
global:
    default-command: version
    initialize: false
    priority: low
`},
		{"hcl", `
"global" = {
    default-command = "version"
    initialize = false
    priority = "low"
}
`},
		{"hcl", `
"global" = {
    default-command = "version"
    initialize = true
}

"global" = {
    initialize = false
    priority = "low"
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			global, err := c.GetGlobalSection()
			require.NoError(t, err)
			assert.Equal(t, "version", global.DefaultCommand)
			assert.Equal(t, false, global.Initialize)
			assert.Equal(t, "low", global.Priority)
			assert.Equal(t, false, global.IONice)
		})
	}
}
