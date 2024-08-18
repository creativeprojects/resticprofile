package remote_test

import (
	"context"
	"testing"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/resticprofile/remote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartAndStopServer(t *testing.T) {
	done := make(chan interface{})
	err := remote.StartServer(done)
	require.NoError(t, err)

	assert.NotEmpty(t, remote.GetPort())

	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	remote.StopServer(ctx)

	<-done
}

func TestClient(t *testing.T) {
	done := make(chan interface{})
	err := remote.StartServer(done)
	require.NoError(t, err)

	assert.NotEmpty(t, remote.GetPort())

	time.Sleep(1 * time.Second)

	client := remote.NewClient(remote.GetPort())
	err = client.LogEntry(clog.LogEntry{
		Level:  clog.LevelWarning,
		Values: []interface{}{"Hello, World!"},
	})
	require.NoError(t, err)

	err = client.Term([]byte("Hello, World!"))
	require.NoError(t, err)

	err = client.Done()
	require.NoError(t, err)

	<-done
}
