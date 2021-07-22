package prom

import (
	"testing"

	"github.com/creativeprojects/resticprofile/shell"
)

func TestSaveTo(t *testing.T) {
	p := NewBackup()
	p.Results(StatusSuccess, shell.Summary{
		BytesAdded: 100,
		BytesTotal: 1000,
	})
	p.SaveTo("test.prom")
}
