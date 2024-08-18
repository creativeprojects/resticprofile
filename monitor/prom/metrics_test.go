package prom

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/monitor"
	"github.com/stretchr/testify/require"
)

func TestSaveSingleBackup(t *testing.T) {
	p := NewMetrics("test", "", "", nil)
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_no_group.prom")
	require.NoError(t, err)
}

func TestSaveSingleBackupWithConfigLabel(t *testing.T) {
	p := NewMetrics("test", "", "", map[string]string{"test_label": "test_value"})
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_add_label.prom")
	require.NoError(t, err)
}

func TestSaveBackupGroup(t *testing.T) {
	p := NewMetrics("test", "group", "", nil)
	p.BackupResults(StatusSuccess, monitor.Summary{
		Duration:   11 * time.Second,
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_group.prom")
	require.NoError(t, err)
}
