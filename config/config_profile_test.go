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

			profiles := c.GetProfileSections()
			assert.NotNil(t, profiles)
			assert.Empty(t, profiles)
		})
	}
}

func TestGetProfileSectionsWithNoProfile(t *testing.T) {
	testData := []testProfileData{
		{"toml", `
[global]
[groups]
`},
		{"json", `{"global":{}, "groups": {}}`},
		{"yaml", `---
global:
groups:
`},
		{"hcl", `
global = {}
groups = {}
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
		{"toml", `
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
		{"json", `{"global":{"something": true},
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
		{"yaml", `---
global:
  something: true
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
		{"hcl", `
global = {
	something = true
}
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
		{"hcl", `
global = {
	something = true
}
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
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			profileSections := c.GetProfileSections()
			assert.NotNil(t, profileSections)
			assert.Len(t, profileSections, 3)

			assert.ElementsMatch(t, []string{"backup"}, profileSections["profile1"], "expected ListA but found ListB")
			assert.ElementsMatch(t, []string{"backup", "snapshots"}, profileSections["profile2"], "expected ListA but found ListB")
			assert.ElementsMatch(t, []string{}, profileSections["profile3"], "expected ListA but found ListB")
		})
	}
}
