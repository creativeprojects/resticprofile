package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigVersion01(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
version = "unknown"
`},
		{"json", `
{
  "version": "unknown"
}`},
		{"yaml", `---
version: unknown
`},
		{"hcl", `
version = "unknown"
`},
		{"toml", `
other = "unknown"
`},
		{"json", `
{
  "other": "unknown"
}`},
		{"yaml", `---
other: unknown
`},
		{"hcl", `
other = "unknown"
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			version := c.GetVersion()
			assert.Equal(t, Version01, version)
		})
	}
}

func TestConfigVersion02(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
version = 2
`},
		{"json", `
{
  "version": 2
}`},
		{"yaml", `---
version: 2
`},
		{"hcl", `
version = 2
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			version := c.GetVersion()
			assert.Equal(t, Version02, version)
		})
	}
}
