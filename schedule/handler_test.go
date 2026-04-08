//go:build !windows

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupExistingBinary(t *testing.T) {
	err := lookupBinary("sh", "sh")
	assert.NoError(t, err)
}

func TestLookupNonExistingBinary(t *testing.T) {
	err := lookupBinary("something", "almost_certain_not_to_be_available")
	assert.Error(t, err)
}
