//go:build !darwin && !windows

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerDefaultOS(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.IsType(t, &HandlerSystemd{}, handler)
}
