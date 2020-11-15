package main

import (
	"testing"

	"github.com/creativeprojects/clog"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	clog.SetTestLog(t)
	defer clog.CloseTestLog()

	err := confirmAndSelfUpdate(true, true, "0.0.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to update binary: file ")
	assert.Contains(t, err.Error(), "resticprofile.test")
	assert.Contains(t, err.Error(), " is not found")
}
