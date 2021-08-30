package config

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

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

func TestStringWithCommaNotConvertedToSlice(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
run-before = "first, second, third"
run-after = ["first", "second", "third"]
`},
		{"json", `
{
  "profile": {
    "run-before": "first, second, third",
    "run-after": ["first", "second", "third"]
  }
}`},
		{"yaml", `---
profile:
    run-before: first, second, third
    run-after: ["first", "second", "third"]
`},
		{"hcl", `
"profile" = {
    run-before = "first, second, third"
    run-after = ["first", "second", "third"]
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			profile, err := c.GetProfile("profile")
			require.NoError(t, err)

			assert.NotNil(t, profile)
			assert.Len(t, profile.RunBefore, 1)
			assert.Len(t, profile.RunAfter, 3)
		})
	}
}

func TestGetIncludes(t *testing.T) {
	config, err := Load(bytes.NewBufferString(`includes=["i1", "i2"]`), "toml")
	require.NoError(t, err)
	assert.Equal(t, config.getIncludes(), []string{"i1", "i2"})

	config, err = Load(bytes.NewBufferString(`includes="inc"`), "toml")
	require.NoError(t, err)
	assert.Equal(t, config.getIncludes(), []string{"inc"})

	config, err = Load(bytes.NewBufferString(`
[includes]
x=0
	`), "toml")
	require.NoError(t, err)
	assert.Nil(t, config.getIncludes())

	config, err = Load(bytes.NewBufferString(`x=0`), "toml")
	require.NoError(t, err)
	assert.Nil(t, config.getIncludes())
}

func TestIncludes(t *testing.T) {
	files := []string{}
	cleanFiles := func() {
		for _, file := range files {
			os.Remove(file)
		}
		files = files[:0]
	}
	defer cleanFiles()

	createFile := func(suffix, content string) string {
		name := ""
		file, err := os.CreateTemp("", "*-"+suffix)
		if err == nil {
			defer file.Close()
			_, err = file.WriteString(content)
			name = file.Name()
			files = append(files, name)
		}
		require.NoError(t, err)
		return name
	}

	mustLoadConfig := func(configFile string) *Config {
		config, err := LoadFile(configFile, "")
		require.NoError(t, err)
		return config
	}

	testID := fmt.Sprintf("%d", time.Now().Unix())

	t.Run("multiple-includes", func(t *testing.T) {
		defer cleanFiles()
		content := fmt.Sprintf(`includes=['*%[1]s.inc.toml','*%[1]s.inc.yaml','*%[1]s.inc.json']`, testID)

		configFile := createFile("profiles.conf", content)
		createFile("d-"+testID+".inc.toml", `[one]`)
		createFile("o-"+testID+".inc.yaml", `two: {}`)
		createFile("j-"+testID+".inc.json", `{"three":{}}`)

		config := mustLoadConfig(configFile)
		assert.True(t, config.IsSet("includes"))
		assert.True(t, config.HasProfile("one"))
		assert.True(t, config.HasProfile("two"))
		assert.True(t, config.HasProfile("three"))
	})

	t.Run("overrides", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile("profiles.conf", `
includes = "*`+testID+`.inc.toml"
[default]
repository = "default-repo"`)

		createFile("override-"+testID+".inc.toml", `
[default]
repository = "overridden-repo"`)

		config := mustLoadConfig(configFile)
		assert.True(t, config.HasProfile("default"))

		profile, err := config.GetProfile("default")
		assert.NoError(t, err)
		assert.Equal(t, NewConfidentialValue("overridden-repo"), profile.Repository)
	})

	t.Run("hcl-includes-only-hcl", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile("profiles.hcl", `includes = "*`+testID+`.inc.*"`)
		createFile("pass-"+testID+".inc.hcl", `one { }`)

		config := mustLoadConfig(configFile)
		assert.True(t, config.HasProfile("one"))

		createFile("fail-"+testID+".inc.toml", `[two]`)
		_, err := LoadFile(configFile, "")
		assert.Error(t, err)
		assert.Regexp(t, ".+ is in hcl format, includes must use the same format", err.Error())
	})

	t.Run("non-hcl-include-no-hcl", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile("profiles.toml", `includes = "*`+testID+`.inc.*"`)
		createFile("pass-"+testID+".inc.toml", `[one]`)

		config := mustLoadConfig(configFile)
		assert.True(t, config.HasProfile("one"))

		createFile("fail-"+testID+".inc.hcl", `one { }`)
		_, err := LoadFile(configFile, "")
		assert.Error(t, err)
		assert.Regexp(t, "hcl format .+ cannot be used in includes from toml", err.Error())
	})
}
