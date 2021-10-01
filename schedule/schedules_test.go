package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseEmptySchedules(t *testing.T) {
	_, err := parseSchedules([]string{})
	assert.NoError(t, err)
}

func TestParseSchedulesWithEmpty(t *testing.T) {
	_, err := parseSchedules([]string{""})
	assert.Error(t, err)
}
