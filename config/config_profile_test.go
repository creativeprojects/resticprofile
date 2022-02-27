package config

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testProfileData struct {
	format string
	config string
}

func TestGetProfileSectionsWithNothing(t *testing.T) {
	testData := []testProfileData{
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

			profiles := c.GetProfileSections()
			assert.NotNil(t, profiles)
			assert.Empty(t, profiles)
		})
	}
}

func TestGetProfileSectionsWithNoProfile(t *testing.T) {
	testData := []testProfileData{
		{FormatTOML, `
includes = ""
[global]
[groups]
`},
		{FormatJSON, `{"global":{}, "groups": {}}`},
		{FormatYAML, `---
includes:
global:
groups:
`},
		{FormatHCL, `
includes = ""
global = {}
groups = {}
`},
		{FormatTOML, `
version = 2
[global]
[groups]
`},
		{FormatJSON, `{"version":2, "global":{}, "groups": {}}`},
		{FormatYAML, `---
version: 2
global:
groups:
`},
		{FormatTOML, `
version = 2
[global]
[groups]
[profiles]
`},
		{FormatJSON, `{"version":2, "global":{}, "groups": {}, "profiles": {}}`},
		{FormatYAML, `---
version: 2
global:
groups:
profiles:
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			profiles := c.GetProfileSections()
			assert.NotNil(t, profiles)
			assert.Empty(t, profiles)
		})
	}
}

func TestGetProfileSections(t *testing.T) {
	testData := []testProfileData{
		{FormatTOML, `includes = "somefile"
[profile1]
[profile1.backup]
source = "/"
[profile2]
[profile2.backup]
source = "/"
[profile2.snapshots]
host = true
[profile3]
some = "value"
[global]
something = true
`},
		{FormatJSON, `{"global":{"something": true},
"includes": ["somefile"],
"profile1": {
	"backup": {"source": "/"}
},
"profile2": {
	"backup": {"source": "/"},
	"snapshots": {"host": true}
},
"profile3": {
	"some": "value"
}
}`},
		{FormatYAML, `---
global:
  something: true
includes: somefile
profile1:
  backup:
    source: "/"
profile2:
  backup:
    source: "/"
  snapshots:
    host: true
profile3:
  some: "value"
`},
		{FormatHCL, `
global = {
	something = true
}
includes = "somefile"
profile1 {
	backup {
      source = "/"
	}
}
profile2 = {
  backup = {
    source = "/"
  }
  snapshots = {
    host = true
  }
}
profile3 = {
  some = "value"
}
`},
		{FormatHCL, `
global = {
	something = true
}
includes = ["somefile"]
profile1 "backup" {
    source = "/"
}
profile2 "backup" {
    source = "/"
}
profile2 "snapshots" {
    host = true
}
profile3 {
  some = "value"
}
`},
		{FormatTOML, `
version = 2
[profiles]
[profiles.profile1]
[profiles.profile1.backup]
source = "/"
[profiles.profile2]
[profiles.profile2.backup]
source = "/"
[profiles.profile2.snapshots]
host = true
[profiles.profile3]
some = "value"
[global]
something = true
`},
		{FormatJSON, `{"version": 2, "global":{"something": true}, "profiles": {
"profile1": {
"backup": {"source": "/"}
},
"profile2": {
"backup": {"source": "/"},
"snapshots": {"host": true}
},
"profile3": {
"some": "value"
}
}}`},
		{FormatYAML, `---
version: 2
global:
  something: true
profiles:
  profile1:
    backup:
      source: "/"
  profile2:
    backup:
      source: "/"
    snapshots:
      host: true
  profile3:
    some: "value"
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			assert.False(t, c.HasProfile("something"))
			assert.True(t, c.HasProfile("profile1"))
			assert.True(t, c.HasProfile("profile2"))
			assert.True(t, c.HasProfile("profile3"))

			profileSections := c.GetProfileSections()
			assert.NotNil(t, profileSections)
			assert.Len(t, profileSections, 3)

			assert.ElementsMatch(t, []string{"backup"}, profileSections["profile1"].Sections)
			assert.ElementsMatch(t, []string{"backup", "snapshots"}, profileSections["profile2"].Sections)
			assert.ElementsMatch(t, []string{}, profileSections["profile3"].Sections)
		})
	}
}
