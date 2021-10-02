package schedule

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupExistingBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	err := lookupBinary("sh", "sh")
	assert.NoError(t, err)
}

func TestLookupNonExistingBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip()
	}
	err := lookupBinary("something", "almost_certain_not_to_be_available")
	assert.Error(t, err)
}
