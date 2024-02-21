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
			assert.Equal(t, []string{"daily"}, schedules["sname"].Schedules)
		})
	}
}

func TestGetScheduleSectionsOnV1(t *testing.T) {
	c := newConfig("toml")
	assert.Panics(t, func() { c.GetScheduleSections() })
}

func TestGetEmptySchedules(t *testing.T) {
	fixtures := []testTemplate{
		{FormatTOML, `version = "1"`},
		{FormatJSON, `{"version": "1"}`},
		{FormatYAML, `version: "1"`},
		{FormatTOML, `version = "2"`},
		{FormatJSON, `{"version": "2"}`},
		{FormatYAML, `version: "2"`},
	}

	for _, testItem := range fixtures {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			schedules, err := c.GetSchedules()
			require.NoError(t, err)
			assert.Empty(t, schedules)
		})
	}
}

func TestGetSchedules(t *testing.T) {
	fixtures := []testTemplate{
		{FormatTOML, `version = "1"
[profile1]
[profile1.backup]
schedule = "daily"
[profile2]
[profile2.backup]
schedule = "weekly"
`},
		{FormatJSON, `{"version": "1", "profile1": {"backup": {"schedule": "daily"}}, "profile2": {"backup": {"schedule": "weekly"}}}`},
		{FormatYAML, `version: "1"
profile1:
  backup:
    schedule: "daily"
profile2:
  backup:
    schedule: "weekly"
`},
		{FormatTOML, `version = "2"
[schedules]
[schedules.schedule1]
profiles="profile1"
schedule="daily"
[schedules.schedule2]
profiles="profile2"
schedule="weekly"
`},
		{FormatJSON, `{"version": "2", "schedules": {"schedule1": {"profiles": "profile1", "schedule": "daily"}, "schedule2": {"profiles": "profile2", "schedule": "weekly"}}}`},
		{FormatYAML, `version: "2"
schedules:
  schedule1:
    profiles: profile1
    schedule: daily
  schedule2:
    profiles: profile2
    schedule: weekly
`},
	}

	for _, testItem := range fixtures {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			schedules, err := c.GetSchedules()
			require.NoError(t, err)
			require.Len(t, schedules, 2)
			for _, schedule := range schedules {
				assertSchedule(t, schedule)
			}
		})
	}
}

func assertSchedule(t *testing.T, schedule *Schedule) {
	t.Helper()

	assert.Len(t, schedule.Profiles, 1)
	assert.Len(t, schedule.Schedules, 1)
	assert.NotEmpty(t, schedule.ProfileName)
}
