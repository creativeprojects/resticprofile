package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigVersion00(t *testing.T) {
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
			assert.Equal(t, Version00, version)
		})
	}
}

func TestConfigVersion1(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
version = 1
`},
		{"json", `
{
  "version": 1
}`},
		{"yaml", `---
version: 1
`},
		{"hcl", `
version = 1
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
