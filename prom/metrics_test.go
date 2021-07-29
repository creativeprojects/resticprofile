package prom

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/require"
)

func TestSaveSingleBackup(t *testing.T) {
	p := NewMetrics("")
	p.BackupResults("test", StatusSuccess, shell.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_no_group.prom")
	require.NoError(t, err)
}

func TestSaveBackupsInGroup(t *testing.T) {
	p := NewMetrics("full-backup")
	p.BackupResults("test1", StatusSuccess, shell.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 1001,
		BytesTotal: 10001,
	})
	p.BackupResults("test2", StatusSuccess, shell.Summary{
		Duration:   time.Duration(12 * time.Second),
		BytesAdded: 1002,
		BytesTotal: 10002,
	})
	err := p.SaveTo("test_group.prom")
	require.NoError(t, err)
}
