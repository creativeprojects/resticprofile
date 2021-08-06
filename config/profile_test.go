package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoProfile(t *testing.T) {
	testConfig := ""
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, profile)
}

func TestEmptyProfile(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
}

func TestNoInitializeValue(t *testing.T) {
	testConfig := `[profile]
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueFalse(t *testing.T) {
	testConfig := `[profile]
initialize = false
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestInitializeValueTrue(t *testing.T) {
	testConfig := `[profile]
initialize = true
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestInheritedInitializeValueTrue(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
inherit = "parent"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, true, profile.Initialize)
}

func TestOverriddenInitializeValueFalse(t *testing.T) {
	testConfig := `[parent]
initialize = true

[profile]
initialize = false
inherit = "parent"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, false, profile.Initialize)
}

func TestUnknownParent(t *testing.T) {
	testConfig := `[profile]
inherit = "parent"
`
	_, err := getProfile("toml", testConfig, "profile")
	assert.Error(t, err)
}

func TestMultiInheritance(t *testing.T) {
	testConfig := `
[grand-parent]
repository = "grand-parent"
first-value = 1
override-value = 1

[parent]
inherit = "grand-parent"
initialize = true
repository = "parent"
second-value = 2
override-value = 2
quiet = true

[profile]
inherit = "parent"
third-value = 3
verbose = true
quiet = false
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)
	assert.Equal(t, "profile", profile.Name)
	assert.Equal(t, "parent", profile.Repository.String())
	assert.Equal(t, true, profile.Initialize)
	assert.Equal(t, int64(1), profile.OtherFlags["first-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["second-value"])
	assert.Equal(t, int64(3), profile.OtherFlags["third-value"])
	assert.Equal(t, int64(2), profile.OtherFlags["override-value"])
	assert.Equal(t, false, profile.Quiet)
	assert.Equal(t, true, profile.Verbose)
}

func TestProfileCommonFlags(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
quiet = true
verbose = false
repository = "test"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetCommonFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "quiet")
	assert.NotContains(flags, "verbose")
	assert.Contains(flags, "repo")
}

func TestProfileOtherFlags(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
bool-true = true
bool-false = false
string = "test"
zero = 0
empty = ""
float = 4.2
int = 42
# comment
array0 = []
array1 = [1]
array2 = ["one", "two"]
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetCommonFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "bool-true")
	assert.NotContains(flags, "bool-false")
	assert.Contains(flags, "string")
	assert.NotContains(flags, "zero")
	assert.NotContains(flags, "empty")
	assert.Contains(flags, "float")
	assert.Contains(flags, "int")
	assert.NotContains(flags, "array0")
	assert.Contains(flags, "array1")
	assert.Contains(flags, "array2")

	assert.Equal([]string{}, flags["bool-true"])
	assert.Equal([]string{"test"}, flags["string"])
	assert.Equal([]string{strconv.FormatFloat(4.2, 'f', -1, 64)}, flags["float"])
	assert.Equal([]string{"42"}, flags["int"])
	assert.Equal([]string{"1"}, flags["array1"])
	assert.Equal([]string{"one", "two"}, flags["array2"])
}

func TestSetRootInProfileUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}
	testConfig := `
[profile]
status-file = "status"
password-file = "key"
lock = "lock"
[profile.backup]
source = ["backup", "root"]
exclude-file = "exclude"
files-from = "include"
exclude = "exclude"
iexclude = "iexclude"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, profile)

	profile.SetRootPath("/wd")
	assert.Equal(t, "status", profile.StatusFile)
	assert.Equal(t, "/wd/key", profile.PasswordFile)
	assert.Equal(t, "/wd/lock", profile.Lock)
	assert.Equal(t, "", profile.CacheDir)
	assert.ElementsMatch(t, []string{"backup", "root"}, profile.GetBackupSource())
	assert.ElementsMatch(t, []string{"/wd/exclude"}, profile.Backup.ExcludeFile)
	assert.ElementsMatch(t, []string{"/wd/include"}, profile.Backup.FilesFrom)
	assert.ElementsMatch(t, []string{"exclude"}, profile.Backup.Exclude)
	assert.ElementsMatch(t, []string{"iexclude"}, profile.Backup.Iexclude)
}

func TestHostInProfile(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
initialize = true
[profile.backup]
host = true
[profile.snapshots]
host = "ConfigHost"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	profile.SetHost("TestHost")

	flags := profile.GetCommandFlags(constants.CommandBackup)
	assert.NotNil(flags)
	assert.Contains(flags, "host")
	assert.Equal([]string{"TestHost"}, flags["host"])

	flags = profile.GetCommandFlags(constants.CommandSnapshots)
	assert.NotNil(flags)
	assert.Contains(flags, "host")
	assert.Equal([]string{"ConfigHost"}, flags["host"])
}

func TestHostInAllSupportedSections(t *testing.T) {
	assert := assert.New(t)

	// Sections supporting "host" flag
	sections := []string{
		constants.CommandBackup,
		constants.CommandForget,
		constants.CommandSnapshots,
		constants.CommandMount,
		constants.SectionConfigurationRetention,
	}

	assertHostIs := func(expectedHost []string, profile *Profile, section string) {
		assert.NotNil(profile)

		var flags = map[string][]string{}
		if section == constants.SectionConfigurationRetention {
			flags = addOtherFlags(flags, profile.Retention.OtherFlags)
		} else {
			flags = profile.GetCommandFlags(section)
		}

		assert.NotNil(flags)
		assert.Contains(flags, "host")
		assert.Equal(expectedHost, flags["host"])
	}

	testConfig := func(section, host string) string {
		return fmt.Sprintf(`
[profile]
initialize = true
[profile.%s]
host = %s
`, section, host)
	}

	for _, section := range sections {
		// Check that host can be set globally
		profile, err := getProfile("toml", testConfig(section, "true"), "profile")
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(profile)

		assertHostIs(emptyStringArray, profile, section)
		profile.SetHost("TestHost")
		assertHostIs([]string{"TestHost"}, profile, section)

		// Ensure host is set only when host value is true
		profile, err = getProfile("toml", testConfig(section, `"OtherTestHost"`), "profile")
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(profile)

		assertHostIs([]string{"OtherTestHost"}, profile, section)
		profile.SetHost("TestHost")
		assertHostIs([]string{"OtherTestHost"}, profile, section)
	}
}

func TestKeepPathInRetention(t *testing.T) {
	assert := assert.New(t)
	root, err := filepath.Abs("/")
	require.NoError(t, err)
	root = filepath.ToSlash(root)
	testConfig := `
[profile]
initialize = true

[profile.backup]
source = "` + root + `"

[profile.retention]
host = false
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetRetentionFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "path")
	assert.Equal([]string{root}, flags["path"])
}

func TestReplacePathInRetention(t *testing.T) {
	assert := assert.New(t)
	testConfig := `
[profile]
initialize = true

[profile.backup]
source = "/some_other_path"

[profile.retention]
path = "/"
`
	profile, err := getProfile("toml", testConfig, "profile")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(profile)

	flags := profile.GetRetentionFlags()
	assert.NotNil(flags)
	assert.Contains(flags, "path")
	assert.Equal([]string{"/"}, flags["path"])
}

func TestForgetCommandFlags(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
initialize = true

[profile.backup]
source = "/"

[profile.forget]
keep-daily = 1
`},
		{"json", `
{
  "profile": {
    "backup": {"source": "/"},
    "forget": {"keep-daily": 1}
  }
}`},
		{"yaml", `---
profile:
  backup:
    source: "/"
  forget:
    keep-daily: 1
`},
		{"hcl", `
"profile" = {
	backup = {
		source = "/"
	}
	forget = {
		keep-daily = 1
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile, err := getProfile(format, testConfig, "profile")
			require.NoError(t, err)

			assert.NotNil(t, profile)
			assert.NotNil(t, profile.Forget)
			assert.NotEmpty(t, profile.Forget.OtherFlags["keep-daily"])
		})
	}
}

func TestSchedules(t *testing.T) {
	assert := assert.New(t)

	testConfig := func(command string, scheduled bool) string {
		schedule := ""
		if scheduled {
			schedule = `schedule = "@hourly"`
		}

		config := `
[profile]
initialize = true

[profile.%s]
%s
`
		return fmt.Sprintf(config, command, schedule)
	}

	sections := NewProfile(nil, "").SchedulableCommands()
	assert.Len(sections, 5)

	for _, command := range sections {
		// Check that schedule is supported
		profile, err := getProfile("toml", testConfig(command, true), "profile")
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(profile)

		config := profile.Schedules()
		assert.Len(config, 1)
		assert.Equal(config[0].commandName, command)
		assert.Len(config[0].schedules, 1)
		assert.Equal(config[0].schedules[0], "@hourly")

		// Check that schedule is optional
		profile, err = getProfile("toml", testConfig(command, false), "profile")
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(profile)
		assert.Empty(profile.Schedules())
	}
}

// schedule is moving from "retention" to "forget" section
// first test: check the schedule works in "forget" section
func TestForgetSchedule(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
initialize = true

[profile.backup]
source = "/"

[profile.forget]
schedule = "weekly"
`},
		{"json", `
{
  "profile": {
    "backup": {"source": "/"},
    "forget": {"schedule": "weekly"}
  }
}`},
		{"yaml", `---
profile:
  backup:
    source: /
  forget:
    schedule: weekly
`},
		{"hcl", `
"profile" = {
	backup = {
		source = "/"
	}
	forget = {
		schedule = "weekly"
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile, err := getProfile(format, testConfig, "profile")
			require.NoError(t, err)

			assert.NotNil(t, profile)
			assert.NotNil(t, profile.Forget)
			assert.NotEmpty(t, profile.Forget.Schedule)
			assert.False(t, profile.HasDeprecatedRetentionSchedule())
		})
	}
}

// schedule is moving from "retention" to "forget" section
// second test: check the schedule deprecation in the "retention" section
func TestRetentionSchedule(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
initialize = true

[profile.backup]
source = "/"

[profile.retention]
schedule = "weekly"
`},
		{"json", `
{
  "profile": {
    "backup": {"source": "/"},
    "retention": {"schedule": "weekly"}
  }
}`},
		{"yaml", `---
profile:
  backup:
    source: /
  retention:
    schedule: weekly
`},
		{"hcl", `
"profile" = {
	backup = {
		source = "/"
	}
	retention = {
		schedule = "weekly"
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile, err := getProfile(format, testConfig, "profile")
			require.NoError(t, err)

			assert.NotNil(t, profile)
			assert.NotNil(t, profile.Retention)
			assert.NotEmpty(t, profile.Retention.Schedule)
			assert.True(t, profile.HasDeprecatedRetentionSchedule())
		})
	}
}

func TestOtherFlags(t *testing.T) {
	testData := []testTemplate{
		{"toml", `
[profile]
other-flag = "1"
[profile.backup]
other-flag-backup = "backup"
[profile.retention]
other-flag-retention = true
[profile.snapshots]
other-flag-snapshots = true
[profile.check]
other-flag-check = true
[profile.forget]
other-flag-forget = true
[profile.prune]
other-flag-prune = true
[profile.mount]
other-flag-mount = true
`},
		{"json", `
{
  "profile": {
	"other-flag": "1",
    "backup": {"other-flag-backup": "backup"},
    "retention": {"other-flag-retention": true},
    "snapshots": {"other-flag-snapshots": true},
    "check": {"other-flag-check": true},
    "forget": {"other-flag-forget": true},
    "prune": {"other-flag-prune": true},
    "mount": {"other-flag-mount": true}
  }
}`},
		{"yaml", `---
profile:
  other-flag: 1
  backup:
    other-flag-backup: backup
  retention:
    other-flag-retention: true
  snapshots:
    other-flag-snapshots: true
  check:
    other-flag-check: true
  forget:
    other-flag-forget: true
  prune:
    other-flag-prune: true
  mount:
    other-flag-mount: true
`},
		{"hcl", `
"profile" = {
	other-flag = 1
	backup = {
		other-flag-backup = "backup"
	}
	retention = {
		other-flag-retention = true
	}
	snapshots = {
		other-flag-snapshots = true
	}
	check = {
		other-flag-check = true
	}
	forget = {
		other-flag-forget = true
	}
	prune = {
		other-flag-prune = true
	}
	mount = {
		other-flag-mount = true
	}
}
`},
	}

	for _, testItem := range testData {
		format := testItem.format
		testConfig := testItem.config
		t.Run(format, func(t *testing.T) {
			profile, err := getProfile(format, testConfig, "profile")
			require.NoError(t, err)

			require.NotNil(t, profile)
			require.NotNil(t, profile.Backup)
			require.NotNil(t, profile.Retention)
			require.NotNil(t, profile.Check)
			require.NotNil(t, profile.Forget)
			require.NotNil(t, profile.Mount)
			require.NotNil(t, profile.Prune)
			require.NotNil(t, profile.Snapshots)

			flags := profile.GetCommonFlags()
			assert.Equal(t, 1, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])

			flags = profile.GetCommandFlags("backup")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			assert.ElementsMatch(t, []string{"backup"}, flags["other-flag-backup"])

			flags = profile.GetRetentionFlags()
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found := flags["other-flag-retention"]
			assert.True(t, found)

			flags = profile.GetCommandFlags("snapshots")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found = flags["other-flag-snapshots"]
			assert.True(t, found)

			flags = profile.GetCommandFlags("check")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found = flags["other-flag-check"]
			assert.True(t, found)

			flags = profile.GetCommandFlags("forget")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found = flags["other-flag-forget"]
			assert.True(t, found)

			flags = profile.GetCommandFlags("prune")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found = flags["other-flag-prune"]
			assert.True(t, found)

			flags = profile.GetCommandFlags("mount")
			assert.Equal(t, 2, len(flags))
			assert.ElementsMatch(t, []string{"1"}, flags["other-flag"])
			_, found = flags["other-flag-mount"]
			assert.True(t, found)
		})
	}
}

func TestMergeFlags(t *testing.T) {
	testData := []struct{ first, second, final map[string][]string }{
		{nil, nil, nil},
		{
			map[string][]string{"key1": {"value"}},
			nil,
			map[string][]string{"key1": {"value"}},
		},
		{
			nil,
			map[string][]string{"key1": {"value"}},
			map[string][]string{"key1": {"value"}},
		},
		{
			map[string][]string{"key1": {"value"}},
			map[string][]string{"key1": {"other", "one"}},
			map[string][]string{"key1": {"other", "one"}},
		},
		{
			map[string][]string{"key1": {"value"}},
			map[string][]string{"key1": nil},
			map[string][]string{"key1": nil},
		},
		{
			map[string][]string{"key1": {"value"}},
			map[string][]string{"key2": {"other", "one"}},
			map[string][]string{"key1": {"value"}, "key2": {"other", "one"}},
		},
	}

	for _, testItem := range testData {
		t.Run("", func(t *testing.T) {
			result := mergeFlags(testItem.first, testItem.second)
			assert.Equal(t, len(testItem.final), len(result))
			for key, value := range testItem.final {
				finalValue, found := result[key]
				assert.True(t, found)
				assert.ElementsMatch(t, value, finalValue)
			}
		})
	}
}
