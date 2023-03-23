//go:build !no_self_update

package main

import (
	"errors"
	"testing"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	err := confirmAndSelfUpdate(true, true, "0.0.1", false)
	assert.Error(t, err)
	assert.Truef(t, errors.Is(err, selfupdate.ErrExecutableNotFoundInArchive), "error returned isn't wrapping %q but is instead: %q", selfupdate.ErrExecutableNotFoundInArchive, err)
	assert.Contains(t, err.Error(), "resticprofile.test")
}
