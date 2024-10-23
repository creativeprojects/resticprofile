package schedule

import (
	"testing"

	"github.com/creativeprojects/resticprofile/platform"
	"github.com/stretchr/testify/assert"
)

func TestLookupExistingBinary(t *testing.T) {
	if platform.IsWindows() {
		t.Skip()
	}
	err := lookupBinary("sh", "sh")
	assert.NoError(t, err)
}

func TestLookupNonExistingBinary(t *testing.T) {
	if platform.IsWindows() {
		t.Skip()
	}
	err := lookupBinary("something", "almost_certain_not_to_be_available")
	assert.Error(t, err)
}
