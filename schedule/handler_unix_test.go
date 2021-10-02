//go:build !darwin && !windows

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerCrond(t *testing.T) {
	handler := NewHandler(SchedulerCrond{})
	assert.IsType(t, &HandlerCrond{}, handler)
}

func TestHandlerSystemd(t *testing.T) {
	handler := NewHandler(SchedulerSystemd{})
	assert.IsType(t, &HandlerSystemd{}, handler)
}

func TestHandlerDefaultOS(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.IsType(t, &HandlerSystemd{}, handler)
}
