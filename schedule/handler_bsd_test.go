//go:build openbsd || netbsd || freebsd

package schedule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerDefaultOS(t *testing.T) {
	handler := NewHandler(SchedulerDefaultOS{})
	assert.IsType(t, &HandlerCrond{}, handler)
}
