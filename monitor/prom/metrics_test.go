package prom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveSingleBackup(t *testing.T) {
	p := NewMetrics("test", "", "", "", nil)
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo(filepath.Join(t.TempDir(), "test_no_group.prom"))
	require.NoError(t, err)
}

func TestSaveSingleBackupWithConfigLabel(t *testing.T) {
	p := NewMetrics("test", "", "", "", map[string]string{"test_label": "test_value"})
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo(filepath.Join(t.TempDir(), "test_add_label.prom"))
	require.NoError(t, err)
}

func TestSaveBackupGroup(t *testing.T) {
	p := NewMetrics("test", "group", "", "", nil)
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo(filepath.Join(t.TempDir(), "test_group.prom"))
	require.NoError(t, err)
}

func readFile(t *testing.T, filename string) string {
	t.Helper()
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	return string(content)
}

func TestSaveBackupSameAsRegistry(t *testing.T) {
	dir := t.TempDir()
	p := NewMetrics("test", "", "1.0", "0.18.0", map[string]string{"test_label": "test_value"})
	p.BackupResults(StatusSuccess, monitor.Summary{Duration: 11 * time.Second, BytesAdded: 100})

	require.NoError(t, p.SaveTo(filepath.Join(dir, "saved.prom")))
	require.NoError(t, prometheus.WriteToTextfile(filepath.Join(dir, "registry.prom"), p.registry))

	saved := readFile(t, filepath.Join(dir, "saved.prom"))
	assert.Equal(t, readFile(t, filepath.Join(dir, "registry.prom")), saved)
	assert.NotContains(t, saved, "resticprofile_check_")
}

func TestCommandResults(t *testing.T) {
	for _, command := range []string{"check", "copy", "forget", "prune"} {
		t.Run(command, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "test.prom")
			p := NewMetrics("test", "", "", "", map[string]string{"repo": "nas"})
			p.CommandResults(command, StatusFailed, 11, monitor.Summary{Duration: 90 * time.Second})
			require.NoError(t, p.SaveTo(filename))

			content := readFile(t, filename)
			labels := `{profile="test",repo="nas"}`
			assert.Contains(t, content, "resticprofile_"+command+"_duration_seconds"+labels+" 90\n")
			assert.Contains(t, content, "resticprofile_"+command+"_exit_code"+labels+" 11\n")
			assert.Contains(t, content, "resticprofile_"+command+"_status"+labels+" 0\n")
			assert.Regexp(t, "resticprofile_"+command+`_time_seconds\{profile="test",repo="nas"\} \d\.\d+e\+09\n`, content)
			assert.NotContains(t, content, "resticprofile_backup_")
		})
	}
}

func TestCommandResultsUnknownCommand(t *testing.T) {
	p := NewMetrics("test", "", "", "", nil)
	assert.False(t, p.HasCommand("unlock"))
	assert.False(t, p.HasCommand("backup"))
	p.CommandResults("unlock", StatusSuccess, 0, monitor.Summary{})

	filename := filepath.Join(t.TempDir(), "test.prom")
	require.NoError(t, p.SaveTo(filename))
	assert.NotContains(t, readFile(t, filename), "unlock")
}

func TestSaveDoesNotOverwriteOtherCommands(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test.prom")

	// each run of resticprofile has its own metrics
	backup := NewMetrics("test", "", "", "", nil)
	backup.BackupResults(StatusSuccess, monitor.Summary{Duration: 11 * time.Second})
	require.NoError(t, backup.SaveTo(filename))

	copyRun := NewMetrics("test", "", "", "", nil)
	copyRun.CommandResults("copy", StatusSuccess, 0, monitor.Summary{Duration: 12 * time.Second})
	require.NoError(t, copyRun.SaveTo(filename))

	check := NewMetrics("test", "", "", "", nil)
	check.CommandResults("check", StatusFailed, 11, monitor.Summary{Duration: 13 * time.Second})
	require.NoError(t, check.SaveTo(filename))

	content := readFile(t, filename)
	assert.Contains(t, content, `resticprofile_backup_duration_seconds{profile="test"} 11`+"\n")
	assert.Contains(t, content, `resticprofile_copy_duration_seconds{profile="test"} 12`+"\n")
	assert.Contains(t, content, `resticprofile_copy_exit_code{profile="test"} 0`+"\n")
	assert.Contains(t, content, `resticprofile_check_duration_seconds{profile="test"} 13`+"\n")
	assert.Contains(t, content, `resticprofile_check_exit_code{profile="test"} 11`+"\n")

	// a new backup run replaces the backup metrics only
	backup = NewMetrics("test", "", "", "", nil)
	backup.BackupResults(StatusFailed, monitor.Summary{Duration: 21 * time.Second})
	require.NoError(t, backup.SaveTo(filename))

	content = readFile(t, filename)
	assert.Contains(t, content, `resticprofile_backup_duration_seconds{profile="test"} 21`+"\n")
	assert.NotContains(t, content, `resticprofile_backup_duration_seconds{profile="test"} 11`)
	assert.Contains(t, content, `resticprofile_copy_duration_seconds{profile="test"} 12`+"\n")
	assert.Contains(t, content, `resticprofile_check_duration_seconds{profile="test"} 13`+"\n")
	assert.Equal(t, 1, strings.Count(content, "# TYPE resticprofile_check_duration_seconds gauge"))

	// the file is still valid
	_, err := loadTextfile(filename)
	require.NoError(t, err)
}

func TestSaveKeepsCommandSeriesWithOtherLabels(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test.prom")

	check := NewMetrics("test", "", "", "", map[string]string{"kind": "quick"})
	check.CommandResults("check", StatusSuccess, 0, monitor.Summary{Duration: 10 * time.Second})
	require.NoError(t, check.SaveTo(filename))

	check = NewMetrics("test", "", "", "", map[string]string{"kind": "full"})
	check.CommandResults("check", StatusFailed, 1, monitor.Summary{Duration: 20 * time.Second})
	require.NoError(t, check.SaveTo(filename))

	check = NewMetrics("test", "", "", "", map[string]string{"kind": "quick"})
	check.CommandResults("check", StatusSuccess, 0, monitor.Summary{Duration: 30 * time.Second})
	require.NoError(t, check.SaveTo(filename))

	content := readFile(t, filename)
	assert.Contains(t, content, `resticprofile_check_duration_seconds{kind="full",profile="test"} 20`+"\n")
	assert.Contains(t, content, `resticprofile_check_duration_seconds{kind="quick",profile="test"} 30`+"\n")
	assert.Equal(t, 2, strings.Count(content, "resticprofile_check_duration_seconds{"))
}

func TestSaveReplacesBuildInfo(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test.prom")

	p := NewMetrics("test", "", "1.0", "0.17.0", nil)
	p.CommandResults("check", StatusSuccess, 0, monitor.Summary{})
	require.NoError(t, p.SaveTo(filename))

	p = NewMetrics("test", "", "2.0", "0.18.0", nil)
	p.CommandResults("copy", StatusSuccess, 0, monitor.Summary{})
	require.NoError(t, p.SaveTo(filename))

	content := readFile(t, filename)
	assert.Contains(t, content, `version="2.0"`)
	assert.Contains(t, content, `version="0.18.0"`)
	assert.NotContains(t, content, `version="1.0"`)
	assert.NotContains(t, content, `version="0.17.0"`)
	assert.Contains(t, content, "resticprofile_check_status")
	assert.Contains(t, content, "resticprofile_copy_status")
}

func TestSaveReplacesInvalidFile(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "test.prom")
	require.NoError(t, os.WriteFile(filename, []byte("not a prometheus file\n"), 0o600))

	p := NewMetrics("test", "", "", "", nil)
	p.CommandResults("check", StatusSuccess, 0, monitor.Summary{})
	require.NoError(t, p.SaveTo(filename))

	content := readFile(t, filename)
	assert.NotContains(t, content, "not a prometheus file")
	assert.Contains(t, content, `resticprofile_check_status{profile="test"} 2`)
}

func TestCommandGatherer(t *testing.T) {
	p := NewMetrics("test", "", "", "", nil)
	p.BackupResults(StatusSuccess, monitor.Summary{})
	p.CommandResults("check", StatusSuccess, 0, monitor.Summary{})
	p.CommandResults("forget", StatusSuccess, 0, monitor.Summary{})

	families, err := p.commandGatherer("check").Gather()
	require.NoError(t, err)
	names := make([]string, 0, len(families))
	for _, family := range families {
		names = append(names, family.GetName())
	}
	assert.ElementsMatch(t, []string{
		"restic_build_info",
		"resticprofile_build_info",
		"resticprofile_check_duration_seconds",
		"resticprofile_check_exit_code",
		"resticprofile_check_status",
		"resticprofile_check_time_seconds",
	}, names)

	families, err = p.commandGatherer("backup").Gather()
	require.NoError(t, err)
	for _, family := range families {
		assert.True(t, isBuildInfo(family.GetName()) || strings.HasPrefix(family.GetName(), "resticprofile_backup_"), family.GetName())
	}
	assert.Len(t, families, 15)
}
