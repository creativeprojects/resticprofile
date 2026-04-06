package write

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncWriter(t *testing.T) {
	buffer := new(bytes.Buffer)
	w := NewAsync(buffer)

	n, err := w.Write([]byte("hello world"))
	require.NoError(t, err)
	assert.Equal(t, 11, n)

	err = w.Close()
	require.NoError(t, err)

	assert.Equal(t, "hello world", buffer.String())
}

func TestAsyncWriteMoreThanChannelSize(t *testing.T) {
	buffer := new(bytes.Buffer)
	w := NewAsync(buffer)

	for range asyncWriterDataChanSize + 1 {
		n, err := w.Write([]byte("aaa"))
		require.NoError(t, err)
		assert.Equal(t, 3, n)
	}

	err := w.Close()
	require.NoError(t, err)

	assert.Equal(t, strings.Repeat("aaa", asyncWriterDataChanSize+1), buffer.String())
}

func TestAsyncWriteInParallelAndClose(t *testing.T) {
	repeat := 10
	buffer := new(bytes.Buffer)
	w := NewAsync(buffer)

	wg := new(sync.WaitGroup)
	for range repeat {
		wg.Go(func() {
			_, _ = w.Write([]byte("aa"))
		})
	}

	require.NoError(t, w.Close())
	wg.Wait()
}

func TestAsyncWriteInParallelAndWaitBeforeClosing(t *testing.T) {
	repeat := 10
	buffer := new(bytes.Buffer)
	w := NewAsync(buffer)

	wg := new(sync.WaitGroup)
	for range repeat {
		wg.Go(func() {
			_, _ = w.Write([]byte("aa"))
		})
	}

	wg.Wait()
	require.NoError(t, w.Close())

	assert.Equal(t, strings.Repeat("aa", repeat), buffer.String())
}

func TestAsyncWriteBigBuffersInParallelAndWaitBeforeClosing(t *testing.T) {
	repeat := 100
	bufferSize := 1024 * 1024
	buffer := new(bytes.Buffer)
	w := NewAsync(buffer)

	wg := new(sync.WaitGroup)
	for range repeat {
		wg.Go(func() {
			buffer := make([]byte, bufferSize)
			_, _ = w.Write(buffer)
		})
	}

	wg.Wait()
	require.NoError(t, w.Close())
	assert.Equal(t, repeat*bufferSize, buffer.Len())
}
