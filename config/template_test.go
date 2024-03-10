package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/creativeprojects/resticprofile/restic"
	"github.com/creativeprojects/resticprofile/util/collect"
	"github.com/creativeprojects/resticprofile/util/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLock(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "tmp\\profile1.lock", profile.Lock)
		return
	}
	assert.Equal(t, "/tmp/profile1.lock", profile.Lock)
}

func TestResolveYear(t *testing.T) {
	testConfig := `
[profile1]
cache-dir = "{{ .Now.Year }}-{{ (.Now.AddDate -1 0 0).Year }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	year := time.Now().Year()
	assert.Equal(t, fmt.Sprintf("%d-%d", year, year-1), profile.CacheDir)
}

func TestResolveSliceValue(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
run-before = ["echo {{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.RunBefore[0])
}

func TestResolveStructThenSliceValue(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
run-before = ["echo {{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.Backup.RunBefore[0])
}

func TestResolveStructThenSliceTwoValues(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
run-before = ["echo {{ .Profile.Name }}", "ls -al"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.Backup.RunBefore, "echo profile1")
}

func TestYamlResolveStructThenSliceValue(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `---
profile1:
  backup:
    run-before:
      - echo {{ .Profile.Name }}
`
	profile, err := getResolvedProfile("yaml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "echo profile1", profile.Backup.RunBefore[0])
}

func TestYamlResolveStructThenSliceTwoValue(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `---
profile1:
  backup:
    run-before:
      - echo {{ .Profile.Name }}
      - ls -al
`
	profile, err := getResolvedProfile("yaml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.Backup.RunBefore, "echo profile1")
}

func TestResolveRemainMap(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
something = "{{ .Profile.Name }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "profile1", profile.OtherFlags["something"])
}

func TestResolveStructThenRemainMap(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.backup]
something = "{{ .Profile.Name }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "profile1", profile.Backup.OtherFlags["something"])
}

func TestResolveInheritanceOfProfileName(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
[profile2]
inherit = "profile1"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile2")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	if runtime.GOOS == "windows" {
		assert.Equal(t, "tmp\\profile2.lock", profile.Lock)
		return
	}
	assert.Equal(t, "/tmp/profile2.lock", profile.Lock)
}

func TestNoInheritanceOfProfileDescription(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
description = "something cool"
[profile2]
inherit = "profile1"
`
	profile1, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile1)

	assert.Equal(t, "something cool", profile1.Description)

	profile2, err := getResolvedProfile("toml", testConfig, "profile2")
	require.NoError(t, err)
	require.NotEmpty(t, profile2)

	assert.Equal(t, "", profile2.Description)
}

func TestResolveSnapshotTag(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.snapshots]
tag = ["test", "{{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"], "profile1")
}

func TestResolveHostname(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	testConfig := `
[profile1]
[profile1.snapshots]
tag = "{{ .Hostname }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"], hostname)
}

func TestResolveGoOSAndArch(t *testing.T) {
	testConfig := `
[profile1.snapshots]
tag = ["{{ .OS }}-{{ .Arch }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	expected := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	assert.Contains(t, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"], expected)
}

func TestResolveEnv(t *testing.T) {
	testConfig := `
[profile1.snapshots]
tag = ["{{ .Env.__NOT_THERE__  | or .Env.__IS_THERE__ }}"]
`
	_ = os.Unsetenv("__NOT_THERE__")
	assert.NoError(t, os.Setenv("__IS_THERE__", "SimpleValue"))

	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"], os.Getenv("__IS_THERE__"))
}

func TestResolveCurrentDir(t *testing.T) {
	testConfig := `
[profile1]
[profile1.snapshots]
tag = "{{ .CurrentDir }}"
`

	currentDir, err := os.Getwd()
	require.NoError(t, err)
	currentDir = filepath.ToSlash(currentDir)

	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, currentDir, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"])
}

func TestResolveTempDir(t *testing.T) {
	testConfig := `
[profile1.snapshots]
tag = "{{ .TempDir }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	tempDir := filepath.ToSlash(os.TempDir())
	assert.Equal(t, tempDir, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"])
}

func TestResolveBinaryDir(t *testing.T) {
	testConfig := `
[profile1]
[profile1.snapshots]
tag = "{{ .BinaryDir }}"
`

	binary, err := os.Executable()
	require.NoError(t, err)
	binaryDir := filepath.ToSlash(filepath.Dir(binary))

	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, binaryDir, profile.OtherSections[constants.CommandSnapshots].OtherFlags["tag"])
}

func TestInheritanceWithTemplates(t *testing.T) {
	// clog.SetTestLog(t)
	// defer clog.CloseTestLog()

	testConfig := `
{{ define "repo" -}}
repository = "/mnt/backup"
{{- end }}

[profile]
{{ template "repo" }}
`

	// try compiling the template manually first
	temp, err := templates.New("").Parse(testConfig)
	require.NoError(t, err)
	buffer := &strings.Builder{}
	err = temp.Execute(buffer, nil)
	require.NoError(t, err)
	assert.Equal(t, "\n\n\n[profile]\nrepository = \"/mnt/backup\"\n", buffer.String())

	profile, err := getResolvedProfile("toml", testConfig, "profile")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/mnt/backup", profile.Repository.String())
}

func TestNestedTemplate(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
{{ define "tags" }}
tag = "{{ .Profile.Name }}"
{{ end }}
[profile]
initialize = true

[profile.backup]
source = "/"
{{ template "tags" . }}

[profile.forget]
keep-daily = 1
{{ template "tags" . }}
`},
		{"json", `
{{ define "tags" }}
"tag": "{{ .Profile.Name }}"
{{ end }}
{
  "profile": {
    "backup": {
		"source": "/",
		{{ template "tags" . }}
	},
    "forget": {
		"keep-daily": 1,
		{{ template "tags" . }}
	}
  }
}`},
		{"yaml", `---
{{ define "tags" }}
    tag: {{ .Profile.Name }}
{{ end }}
profile:
  backup:
    source: "/"
    {{ template "tags" . }}
  forget:
    keep-daily: 1
    {{ template "tags" . }}
`},
		{"hcl", `
{{ define "tags" }}
    tag = "{{ .Profile.Name }}"
{{ end }}
"profile" = {
	backup = {
		source = "/"
		{{ template "tags" . }}
	}
	forget = {
		keep-daily = 1
		{{ template "tags" . }}
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile, err := getResolvedProfile(format, testConfig, "profile")
			require.NoError(t, err)

			assert.NotNil(t, profile)
			assert.NotNil(t, profile.Backup)
			assert.Contains(t, profile.Backup.OtherFlags["tag"], "profile")
			assert.NotNil(t, profile.Forget)
			assert.Contains(t, profile.Forget.OtherFlags["tag"], "profile")
		})
	}
}

func TestInfoData(t *testing.T) {
	data := NewTemplateInfoData(restic.AnyVersion)

	assert.Equal(t, time.Now().Truncate(time.Minute), data.Now.Truncate(time.Minute))
	assert.NotEmpty(t, data.TempDir)
	assert.NotEmpty(t, data.Env)
	assert.NotEmpty(t, data.CurrentDir)
	assert.NotEmpty(t, data.Hostname)
	assert.NotEmpty(t, data.BinaryDir)

	assert.IsType(t, new(profileInfo), data.Profile)
	assert.Equal(t, NewGlobalInfo(), data.Global)
	assert.Equal(t, NewGroupInfo(), data.Group)
	assert.Equal(t, restic.KnownVersions(), data.KnownResticVersions)

	assert.NotEmpty(t, data.ProfileSections())

	t.Run("NestedSections", func(t *testing.T) {
		sections := data.NestedSections()
		assert.NotEmpty(t, sections)
		assert.Subset(t, collect.From(sections, SectionInfo.Name), []string{
			"ScheduleBaseConfig",
			"ScheduleConfig",
			"SendMonitoringHeader",
			"SendMonitoringSection",
			"StreamErrorSection",
		})
	})

	funcs := data.GetFuncs()
	props := collect.From(data.Profile.Properties(), data.Profile.PropertyInfo)
	assert.Len(t, funcs, 3)
	assert.NotEmpty(t, props)

	t.Run("properties", func(t *testing.T) {
		fn, ok := funcs["properties"].(func(set PropertySet) []PropertyInfo)
		require.True(t, ok)
		assert.Equal(t, props, fn(data.Profile))
	})

	t.Run("own/restic", func(t *testing.T) {
		own, ok := funcs["own"].(func(p []PropertyInfo) []PropertyInfo)
		require.True(t, ok)
		notOwn, ok := funcs["restic"].(func(p []PropertyInfo) []PropertyInfo)
		require.True(t, ok)

		ownProps := own(props)
		notOwnProps := notOwn(props)

		assert.NotEqual(t, props, ownProps)
		assert.NotEqual(t, props, notOwnProps)
		assert.Subset(t, props, ownProps)
		assert.Subset(t, props, notOwnProps)
		assert.ElementsMatch(t, props, append(notOwnProps, ownProps...))

		for _, prop := range ownProps {
			assert.NotContains(t, notOwnProps, prop)
			assert.Contains(t, props, prop)
			assert.False(t, prop.IsOption())
		}

		for _, prop := range notOwnProps {
			assert.NotContains(t, ownProps, prop)
			assert.Contains(t, props, prop)
			assert.True(t, prop.IsOption())
		}
	})
}
