package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetScheduleSections(t *testing.T) {
	testData := []testTemplate{
		{FormatTOML, `
version = 2
[schedules]
[schedules.sname]
profiles="value"
schedule="daily"
`},
		{FormatJSON, `
{
  "version": 2,
  "schedules": {
    "sname": {
      "profiles": "value",
      "schedule": "daily"
    }
  }
}`},
		{FormatYAML, `---
version: 2
schedules:
  sname:
    profiles: value
    schedule: daily
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			schedules, err := c.GetScheduleSections()
			require.NoError(t, err)
			assert.NotEmpty(t, schedules)
			assert.Equal(t, []string{"value"}, schedules["sname"].Profiles)
			assert.Equal(t, []string{"daily"}, schedules["sname"].Schedule)
		})
	}
}
