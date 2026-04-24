//go:build !windows && !darwin && !openbsd && !netbsd && !freebsd

package schedule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystemdInit(t *testing.T) {
	handler := (NewHandler(SchedulerSystemd{}))
	err := handler.Init()
	defer handler.Close()
	require.NoError(t, err)
}
