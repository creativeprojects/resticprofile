//go:build !windows

package schedule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrondInit(t *testing.T) {
	handler := (NewHandler(SchedulerCrond{}))
	err := handler.Init()
	defer handler.Close()
	require.NoError(t, err)
}
