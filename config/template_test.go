package config

import (
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLock(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/tmp/profile1.lock", profile.Lock)
}

func TestResolveYear(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
cache-dir = "{{ .Now.Year }}"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, strconv.Itoa(time.Now().Year()), profile.CacheDir)
}

func TestResolveSliceValue(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

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
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
lock = "/tmp/{{ .Profile.Name }}.lock"
[profile2]
inherit = "profile1"
`
	profile, err := getResolvedProfile("toml", testConfig, "profile2")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/tmp/profile2.lock", profile.Lock)
}

func TestResolveSnapshotTag(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
[profile1]
[profile1.snapshots]
tag = ["test", "{{ .Profile.Name }}"]
`
	profile, err := getResolvedProfile("toml", testConfig, "profile1")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Contains(t, profile.Snapshots["tag"], "profile1")
}

func TestInheritanceWithTemplates(t *testing.T) {
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	testConfig := `
{{ define "repo" -}}
repository = "/mnt/backup"
{{- end }}

[profile]
{{ template "repo" }}
`

	// try compiling the template manually first
	temp, err := template.New("").Parse(testConfig)
	require.NoError(t, err)
	buffer := &strings.Builder{}
	err = temp.Execute(buffer, nil)
	require.NoError(t, err)
	assert.Equal(t, "\n\n\n[profile]\nrepository = \"/mnt/backup\"\n", buffer.String())

	profile, err := getResolvedProfile("toml", testConfig, "profile")
	require.NoError(t, err)
	require.NotEmpty(t, profile)

	assert.Equal(t, "/mnt/backup", profile.Repository)
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
			assert.Contains(t, profile.Forget["tag"], "profile")
		})
	}
}
