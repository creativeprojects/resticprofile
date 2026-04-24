package schedule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLaunchdInit(t *testing.T) {
	handler := (NewHandler(SchedulerLaunchd{}))
	err := handler.Init()
	defer handler.Close()
	require.NoError(t, err)
}
