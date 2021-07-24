package prom

import (
	"testing"
	"time"

	"github.com/creativeprojects/resticprofile/shell"
	"github.com/stretchr/testify/require"
)

func TestSaveToNoGroup(t *testing.T) {
	p := NewBackup("")
	p.Results("test", StatusSuccess, shell.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_no_group.prom")
	require.NoError(t, err)
}

func TestSaveToWithGroup(t *testing.T) {
	p := NewBackup("full-backup")
	p.Results("test", StatusSuccess, shell.Summary{
		Duration:   time.Duration(11 * time.Second),
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	err := p.SaveTo("test_group.prom")
	require.NoError(t, err)
}
