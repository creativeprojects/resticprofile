//go:build !no_self_update

package main

import (
	"testing"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/resticprofile/term"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("don't run update in short mode")
	}
	// can't run in Parallel because it changes the default logger
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	err := confirmAndSelfUpdate(term.NewTerminal(), true, true, "0.0.1", false)
	require.ErrorIsf(t, err, selfupdate.ErrExecutableNotFoundInArchive, "error returned isn't wrapping %q but is instead: %q", selfupdate.ErrExecutableNotFoundInArchive, err)
	assert.Contains(t, err.Error(), "resticprofile.test")
}
