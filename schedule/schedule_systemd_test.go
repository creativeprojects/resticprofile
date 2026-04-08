//go:build !windows && !darwin

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
