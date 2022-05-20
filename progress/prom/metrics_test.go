package prom

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/progress"
	"github.com/stretchr/testify/require"
)

func TestSaveSingleBackup(t *testing.T) {
	p := NewMetrics("", "", nil)
	p.BackupResults("test", StatusSuccess, progress.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_no_group.prom")
	require.NoError(t, err)
}

func TestSaveSingleBackupWithConfigLabel(t *testing.T) {
	p := NewMetrics("", "", map[string]string{"test_label": "test_value"})
	p.BackupResults("test", StatusSuccess, progress.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_add_label.prom")
	require.NoError(t, err)
}

func TestSaveBackupsInGroup(t *testing.T) {
	p := NewMetrics("full-backup", "", nil)
	p.BackupResults("test1", StatusSuccess, progress.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 1001,
		BytesTotal: 10001,
	})
	p.BackupResults("test2", StatusSuccess, progress.Summary{
		Duration:   time.Duration(12 * time.Second),
		BytesAdded: 1002,
		BytesTotal: 10002,
	})
	err := p.SaveTo("test_group.prom")
	require.NoError(t, err)
}
