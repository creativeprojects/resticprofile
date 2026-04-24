package schedule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWindowsInit(t *testing.T) {
	handler := (NewHandler(SchedulerWindows{}))
	err := handler.Init()
	defer handler.Close()
	require.NoError(t, err)
}
