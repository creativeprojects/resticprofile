//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/creativeprojects/resticprofile/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetRootInProfileUnix(t *testing.T) {
	runForVersions(t, func(t *testing.T, version, prefix string) {
		t.Helper()
		testConfig := version + `
[` + prefix + `profile]
base-dir = "~"
status-file = "status"
prometheus-save-to-file = "prom"
repository = "local-repo"
password-file = "key"
lock = "lock"
[` + prefix + `profile.backup]
source = ["backup", "root"]
exclude-file = "exclude"
iexclude-file = "iexclude"
files-from = "include"
files-from-raw = "include-raw"
files-from-verbatim = "include-verbatim"
exclude = "exclude"
iexclude = "iexclude"
[` + prefix + `profile.copy]
from-password-file = "key"
[` + prefix + `profile.dump]
password-file = "key"
[` + prefix + `profile.init]
from-repository-file = "key"
from-password-file = "key"
`
		profile, err := getProfile("toml", testConfig, "profile", "")
		require.NoError(t, err)
		require.NotNil(t, profile)

		homeDir, err := os.UserHomeDir()
		require.NoError(t, err)

		profile.ResolveConfiguration()
		assert.Equal(t, homeDir, profile.BaseDir)
		assert.Equal(t, "local-repo", profile.Repository.Value())

		profile.SetRootPath("/wd")
		assert.Equal(t, "status", profile.StatusFile)
		assert.Equal(t, "prom", profile.PrometheusSaveToFile)
		assert.Equal(t, "/wd/key", profile.PasswordFile)
		assert.Equal(t, "/wd/lock", profile.Lock)
		assert.Equal(t, "", profile.CacheDir)
		assert.ElementsMatch(t, []string{
			filepath.Join(homeDir, "backup"),
			filepath.Join(homeDir, "root"),
		}, profile.GetBackupSource())
		assert.ElementsMatch(t, []string{"/wd/exclude"}, profile.Backup.ExcludeFile)
		assert.ElementsMatch(t, []string{"/wd/iexclude"}, profile.Backup.IexcludeFile)
		assert.ElementsMatch(t, []string{"/wd/include"}, profile.Backup.FilesFrom)
		assert.ElementsMatch(t, []string{"/wd/include-raw"}, profile.Backup.FilesFromRaw)
		assert.ElementsMatch(t, []string{"/wd/include-verbatim"}, profile.Backup.FilesFromVerbatim)
		assert.ElementsMatch(t, []string{"exclude"}, profile.Backup.Exclude)
		assert.ElementsMatch(t, []string{"iexclude"}, profile.Backup.Iexclude)
		assert.Equal(t, "/wd/key", profile.Copy.FromPasswordFile)
		assert.Equal(t, []string{"/wd/key"}, profile.OtherSections[constants.CommandDump].OtherFlags["password-file"])
		assert.Equal(t, "/wd/key", profile.Init.FromPasswordFile)
		assert.Equal(t, "/wd/key", profile.Init.FromRepositoryFile)
	})
}

func TestSetRootPathOnMonitoringSections(t *testing.T) {
	sections := SendMonitoringSections{
		SendBefore: []SendMonitoringSection{
			{BodyTemplate: "file"},
		},
		SendAfter: []SendMonitoringSection{
			{BodyTemplate: "file"},
			{BodyTemplate: "file"},
		},
		SendAfterFail: []SendMonitoringSection{
			{BodyTemplate: "file"},
			{BodyTemplate: "file"},
		},
		SendFinally: []SendMonitoringSection{
			{BodyTemplate: "file"},
		},
	}

	sections.setRootPath(nil, "root")
	assert.Equal(t, "root/file", sections.SendBefore[0].BodyTemplate)

	assert.Equal(t, "root/file", sections.SendAfter[0].BodyTemplate)
	assert.Equal(t, "root/file", sections.SendAfter[1].BodyTemplate)

	assert.Equal(t, "root/file", sections.SendAfterFail[0].BodyTemplate)
	assert.Equal(t, "root/file", sections.SendAfterFail[1].BodyTemplate)

	assert.Equal(t, "root/file", sections.SendFinally[0].BodyTemplate)
}
