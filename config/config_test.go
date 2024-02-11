package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/util/maybe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTemplate struct {
	format string
	config string
}

func TestIsSet(t *testing.T) {
	testData := []testTemplate{
		{FormatTOML, `
[first]
[first.second]
key="value"
`},
		{FormatJSON, `
{
  "first": {
    "second": {
		"key": "value"
	}
  }
}`},
		{FormatYAML, `---
first:
  second:
    key: value
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(testConfig), format)
			require.NoError(t, err)

			assert.True(t, c.IsSet("first"))
			assert.True(t, c.IsSet("first", "second"))
			assert.True(t, c.IsSet("first", "second", "key"))
		})
	}
}

func TestGetGlobal(t *testing.T) {
	testData := []testTemplate{
		{FormatTOML, `
[global]
priority = "low"
default-command = "version"
# initialize a repository if none exist at location
initialize = false
`},
		{FormatJSON, `
{
  "global": {
    "default-command": "version",
    "initialize": false,
    "priority": "low"
  }
}`},
		{FormatYAML, `---
global:
    default-command: version
    initialize: false
    priority: low
`},
		{FormatHCL, `
"global" = {
    default-command = "version"
    initialize = false
    priority = "low"
}
`},
		{FormatHCL, `
"global" = {
    default-command = "version"
    initialize = true
}

"global" = {
    initialize = false
    priority = "low"
}
`},
		{FormatTOML, `
version = 2
[global]
priority = "low"
default-command = "version"
# initialize a repository if none exist at location
initialize = false
`},
		{FormatJSON, `
{
  "version": 2,
  "global": {
    "default-command": "version",
    "initialize": false,
    "priority": "low"
  }
}`},
		{FormatYAML, `---
version: 2
global:
    default-command: version
    initialize: false
    priority: low
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

func TestHCLFormatOnlyInVersion01(t *testing.T) {
	testConfig := `
version = 2
"global" = {
    default-command = "version"
    initialize = false
    priority = "low"
}`
	c, err := Load(bytes.NewBufferString(testConfig), FormatHCL)
	require.NoError(t, err)
	_, err = c.GetGlobalSection()
	assert.EqualError(t, err, "HCL format is not supported in version 2, please use version 1 or another file format")
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

func TestBoolPointer(t *testing.T) {
	fixtures := []struct {
		testTemplate
		continueOnError maybe.Bool
	}{
		{
			testTemplate: testTemplate{
				format: FormatTOML,
				config: `
version = 2
[groups]
[groups.groupname]
profiles = []
`,
			},
			continueOnError: maybe.Bool{},
		},
		{
			testTemplate: testTemplate{
				format: FormatYAML,
				config: `
version: 2
groups:
  groupname:
    profiles: []
`,
			},
			continueOnError: maybe.Bool{},
		},
		{
			testTemplate: testTemplate{
				format: FormatJSON,
				config: `
{
	"version": 2,
	"groups": {
		"groupname":{
			"profiles": []
		}
	}
}
`,
			},
			continueOnError: maybe.Bool{},
		},
		{
			testTemplate: testTemplate{
				format: FormatTOML,
				config: `
version = 2
[groups]
[groups.groupname]
profiles = []
continue-on-error = true
`,
			},
			continueOnError: maybe.True(),
		},
		{
			testTemplate: testTemplate{
				format: FormatYAML,
				config: `
version: 2
groups:
 groupname:
  profiles: []
  continue-on-error: true
`,
			},
			continueOnError: maybe.True(),
		},
		{
			testTemplate: testTemplate{
				format: FormatJSON,
				config: `
{
	"version": 2,
	"groups": {
		"groupname":{
			"profiles": [],
			"continue-on-error": true
		}
	}
}
`,
			},
			continueOnError: maybe.True(),
		},
		{
			testTemplate: testTemplate{
				format: FormatTOML,
				config: `
version = 2
[groups]
[groups.groupname]
profiles = []
continue-on-error = false
`,
			},
			continueOnError: maybe.False(),
		},
		{
			testTemplate: testTemplate{
				format: FormatYAML,
				config: `
version: 2
groups:
  groupname:
    profiles: []
    continue-on-error: false
`,
			},
			continueOnError: maybe.False(),
		},
		{
			testTemplate: testTemplate{
				format: FormatJSON,
				config: `
{
	"version": 2,
	"groups": {
		"groupname":{
			"profiles": [],
			"continue-on-error": false
		}
	}
}
`,
			},
			continueOnError: maybe.False(),
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.format, func(t *testing.T) {
			c, err := Load(bytes.NewBufferString(fixture.config), fixture.format)
			require.NoError(t, err)

			group, err := c.GetProfileGroup("groupname")
			require.NoError(t, err)

			assert.Equal(t, fixture.continueOnError, group.ContinueOnError)
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

	createFile := func(t *testing.T, suffix, content string) string {
		t.Helper()
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

	mustLoadConfig := func(t *testing.T, configFile string) *Config {
		t.Helper()
		config, err := LoadFile(configFile, "")
		require.NoError(t, err)
		return config
	}

	testID := fmt.Sprintf("%d", time.Now().Unix())

	t.Run("multiple-includes", func(t *testing.T) {
		defer cleanFiles()
		content := fmt.Sprintf(`includes=['*%[1]s.inc.toml','*%[1]s.inc.yaml','*%[1]s.inc.json']`, testID)

		configFile := createFile(t, "profiles.conf", content)
		createFile(t, "d-"+testID+".inc.toml", "[one]\nk='v'")
		createFile(t, "o-"+testID+".inc.yaml", `two: { k: v }`)
		createFile(t, "j-"+testID+".inc.json", `{"three":{ "k": "v" }}`)

		config := mustLoadConfig(t, configFile)
		assert.True(t, config.IsSet("includes"))
		assert.True(t, config.HasProfile("one"))
		assert.True(t, config.HasProfile("two"))
		assert.True(t, config.HasProfile("three"))
	})

	t.Run("overrides", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile(t, "profiles.conf", `
includes = "*`+testID+`.inc.toml"
[default]
repository = "default-repo"`)

		createFile(t, "override-"+testID+".inc.toml", `
[default]
repository = "overridden-repo"`)

		config := mustLoadConfig(t, configFile)
		assert.True(t, config.HasProfile("default"))

		profile, err := config.GetProfile("default")
		require.NoError(t, err)
		assert.Equal(t, NewConfidentialValue("overridden-repo"), profile.Repository)
	})

	t.Run("mixins", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile(t, "profiles.conf", `
version = 2
includes = "*`+testID+`.inc.toml"
[profiles.default]
use = "another-run-before"
run-before = "default-before"`)

		createFile(t, "mixin-"+testID+".inc.toml", `
[mixins.another-run-before]
"run-before..." = "another-run-before"
[mixins.another-run-before2]
"run-before..." = "another-run-before2"`)

		createFile(t, "mixin-use-"+testID+".inc.toml", `
[profiles.default]
use = "another-run-before2"`)

		config := mustLoadConfig(t, configFile)
		assert.True(t, config.HasProfile("default"))

		profile, err := config.GetProfile("default")
		require.NoError(t, err)
		assert.Equal(t, []string{"default-before", "another-run-before", "another-run-before2"}, profile.RunBefore)
	})

	t.Run("hcl-includes-only-hcl", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile(t, "profiles.hcl", `includes = "*`+testID+`.inc.*"`)
		createFile(t, "pass-"+testID+".inc.hcl", `one { }`)

		config := mustLoadConfig(t, configFile)
		assert.True(t, config.HasProfile("one"))

		createFile(t, "fail-"+testID+".inc.toml", `[two]`)
		_, err := LoadFile(configFile, "")
		assert.Error(t, err)
		assert.Regexp(t, ".+ is in hcl format, includes must use the same format", err.Error())
	})

	t.Run("non-hcl-include-no-hcl", func(t *testing.T) {
		defer cleanFiles()

		configFile := createFile(t, "profiles.toml", `includes = "*`+testID+`.inc.*"`)
		createFile(t, "pass-"+testID+".inc.toml", "[one]\nk='v'")

		config := mustLoadConfig(t, configFile)
		assert.True(t, config.HasProfile("one"))

		createFile(t, "fail-"+testID+".inc.hcl", `one { }`)
		_, err := LoadFile(configFile, "")
		assert.Error(t, err)
		assert.Regexp(t, "hcl format .+ cannot be used in includes from toml", err.Error())
	})

	t.Run("cannot-load-different-versions", func(t *testing.T) {
		defer cleanFiles()
		content := fmt.Sprintf(`includes=['*%s.inc.json']`, testID)

		configFile := createFile(t, "profiles.conf", content)
		createFile(t, "a-"+testID+".inc.json", `{"version": 2, "profiles": {"one":{}}}`)
		createFile(t, "b-"+testID+".inc.json", `{"two":{}}`)

		_, err := LoadFile(configFile, "")
		assert.ErrorContains(t, err, "cannot include different versions of the configuration file")
	})

	t.Run("cannot-load-different-versions", func(t *testing.T) {
		defer cleanFiles()
		content := fmt.Sprintf(`{"version": 2, "includes":["*%s.inc.json"]}`, testID)

		configFile := createFile(t, "profiles.json", content)
		createFile(t, "c-"+testID+".inc.json", `{"version": 1, "two":{}}`)
		createFile(t, "d-"+testID+".inc.json", `{"profiles": {"one":{}}}`)

		_, err := LoadFile(configFile, "")
		assert.ErrorContains(t, err, "cannot include different versions of the configuration file")
	})
}

func TestGetProfiles(t *testing.T) {
	var fixtures = []struct {
		format  string
		content string
	}{
		{
			FormatTOML, `[default]
repository="1"
[default.snapshots]
tags=true
[default.check]
tags=true
[one]
inherit="default"
[two]
repository="2"
`,
		},
		{
			FormatYAML, `---
default:
  repository: "1"
  snapshots:
    tags: true
  check:
    tags: true
one:
  inherit: "default"
two:
  repository: "2"
`,
		},
		{
			FormatJSON, `{"default":{
  "repository":"1",
  "snapshots":{
    "tags":"true"
  },
  "check":{
    "tags":"true"
  }
},
"one":{
  "inherit":"default"
},
"two":{
  "repository":"2"
}
}`,
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.format, func(t *testing.T) {
			buffer := bytes.NewBufferString(fixture.content)
			cfg, err := Load(buffer, fixture.format)
			require.NoError(t, err)
			assert.NotNil(t, cfg)

			configs := cfg.GetProfiles()
			assert.Len(t, configs, 3)

			profile, ok := configs["default"]
			require.True(t, ok)
			require.NotNil(t, profile)

			profile, ok = configs["one"]
			require.True(t, ok)
			require.NotNil(t, profile)

			commands := profile.DefinedCommands()
			assert.ElementsMatch(t, []string{"check", "snapshots"}, commands)

			profile, ok = configs["two"]
			require.True(t, ok)
			require.NotNil(t, profile)

			assert.Empty(t, profile.DefinedCommands())
		})
	}
}

func TestGetSchedules(t *testing.T) {
	content := `---
profile:
  backup:
    schedule: daily
`
	buffer := bytes.NewBufferString(content)
	cfg, err := Load(buffer, FormatYAML)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	schedules, err := cfg.GetSchedules()
	require.NoError(t, err)
	assert.Len(t, schedules, 1)

	assert.Equal(t, []string{"daily"}, schedules[0].Schedules)
}

func TestRegressionInheritanceListMerging(t *testing.T) {
	load := func(content string) *Config {
		buffer := bytes.NewBufferString(content)
		cfg, err := Load(buffer, FormatYAML)
		require.NoError(t, err)
		require.NotNil(t, cfg)
		return cfg
	}

	t.Run("Version1", func(t *testing.T) {
		config := load(`---
base:
  run-before: ["base-first", "base-second", "base-third"]
profile:
  inherit: base
  run-before: [null, "profile-before"]
`)
		profile, err := config.GetProfile("profile")
		require.NoError(t, err)
		assert.Equal(t, []string{"base-first", "profile-before", "base-third"}, profile.RunBefore)
	})

	t.Run("Version2", func(t *testing.T) {
		config := load(`---
version: 2
profiles:
  base:
    run-before: ["base-first", "base-second", "base-third"]
  profile:
    inherit: base
    run-before: [null, "profile-before"]
`)
		profile, err := config.GetProfile("profile")
		require.NoError(t, err)
		assert.Equal(t, []string{"", "profile-before"}, profile.RunBefore)
	})
}

func TestRequireVersionAssertions(t *testing.T) {
	t.Run("v1", func(t *testing.T) {
		c := newConfig("toml")
		assert.NotPanics(t, func() { c.requireVersion(Version01) })
		assert.NotPanics(t, func() { c.requireMinVersion(Version01) })
		assert.Panics(t, func() { c.requireVersion(Version02) })
		assert.Panics(t, func() { c.requireMinVersion(Version02) })
	})

	t.Run("v2", func(t *testing.T) {
		c := newConfig("toml")
		c.viper.Set("version", 2)
		assert.NotPanics(t, func() { c.requireVersion(Version02) })
		assert.NotPanics(t, func() { c.requireMinVersion(Version01) })
		assert.NotPanics(t, func() { c.requireMinVersion(Version02) })
		assert.Panics(t, func() { c.requireVersion(Version01) })
	})
}

// This is a simple fuzzing test on configuration file to see if resticprofile can crash loading a funny file
// I wanted to give it a try but it might just be useless
func FuzzConfigTOML(f *testing.F) {
	examples, err := os.ReadDir("../examples")
	require.NoError(f, err)

	// use files in example dir as seeds
	for _, example := range examples {
		if strings.HasSuffix(example.Name(), ".conf") || strings.HasSuffix(example.Name(), ".toml") {
			func() {
				file, err := os.Open(filepath.Join("../examples", example.Name()))
				require.NoError(f, err)
				defer file.Close()

				buffer := &bytes.Buffer{}
				_, err = io.Copy(buffer, file)
				require.NoError(f, err)
				f.Add(buffer.Bytes())
			}()
		}
	}

	f.Fuzz(func(t *testing.T, in []byte) {
		reader := bytes.NewReader(in)
		// we just want to detect a panic, so we don't check the returned values
		if cfg, err := Load(reader, "toml"); cfg != nil && err == nil {
			_ = cfg.GetProfiles()
		}
	})
}
