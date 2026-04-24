package util

import (
	"bytes"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// trackingWriteCloser is a bytes.Buffer that also tracks Close and Flush calls.
type trackingWriteCloser struct {
	bytes.Buffer
	closed   bool
	flushed  bool
	closeErr error
	flushErr error
}

func (t *trackingWriteCloser) Close() error {
	t.closed = true
	return t.closeErr
}

func (t *trackingWriteCloser) Flush() error {
	t.flushed = true
	return t.flushErr
}

// --- FlushWriter tests ---

func TestFlushWriterFlushable(t *testing.T) {
	tw := &trackingWriteCloser{}
	flushable, err := FlushWriter(tw)
	assert.True(t, flushable)
	assert.NoError(t, err)
	assert.True(t, tw.flushed)
}

func TestFlushWriterNotFlushable(t *testing.T) {
	var buf bytes.Buffer
	flushable, err := FlushWriter(&buf)
	assert.False(t, flushable)
	assert.NoError(t, err)
}

func TestFlushWriterFlushError(t *testing.T) {
	sentinel := errors.New("flush error")
	tw := &trackingWriteCloser{flushErr: sentinel}
	flushable, err := FlushWriter(tw)
	assert.True(t, flushable)
	assert.Equal(t, sentinel, err)
}

// --- syncWriter tests ---

func TestSyncWriterWrite(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	n, err := w.Write([]byte("hello"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", buf.String())
}

func TestSyncWriterMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	for _, s := range []string{"foo", "bar", "baz"} {
		_, err := w.Write([]byte(s))
		require.NoError(t, err)
	}
	assert.Equal(t, "foobarbaz", buf.String())
}

func TestSyncWriterLocked(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	err := w.Locked(func(inner *bytes.Buffer) error {
		_, err := inner.WriteString("locked")
		return err
	})
	assert.NoError(t, err)
	assert.Equal(t, "locked", buf.String())
}

func TestSyncWriterLockedPropagatesError(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	sentinel := errors.New("locked error")
	err := w.Locked(func(_ *bytes.Buffer) error { return sentinel })
	assert.Equal(t, sentinel, err)
}

func TestSyncWriterFlushWithFlusher(t *testing.T) {
	tw := &trackingWriteCloser{}
	w := NewSyncWriter(tw)

	// syncWriter exposes Flush via a type assertion at call site
	f, ok := w.(Flusher)
	require.True(t, ok)
	err := f.Flush()
	assert.NoError(t, err)
	assert.True(t, tw.flushed)
}

func TestSyncWriterFlushError(t *testing.T) {
	sentinel := errors.New("flush error")
	tw := &trackingWriteCloser{flushErr: sentinel}
	w := NewSyncWriter(tw)

	err := w.(Flusher).Flush()
	assert.Equal(t, sentinel, err)
}

func TestSyncWriterFlushNonFlusher(t *testing.T) {
	// bytes.Buffer is not a Flusher; Flush should be a no-op
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	err := w.(Flusher).Flush()
	assert.NoError(t, err)
}

func TestSyncWriterCloseWithCloser(t *testing.T) {
	tw := &trackingWriteCloser{}
	w := NewSyncWriter(tw)

	type closer interface{ Close() error }
	err := w.(closer).Close()
	assert.NoError(t, err)
	assert.True(t, tw.closed)
}

func TestSyncWriterCloseError(t *testing.T) {
	sentinel := errors.New("close error")
	tw := &trackingWriteCloser{closeErr: sentinel}
	w := NewSyncWriter(tw)

	type closer interface{ Close() error }
	err := w.(closer).Close()
	assert.Equal(t, sentinel, err)
}

func TestSyncWriterCloseNonCloser(t *testing.T) {
	// bytes.Buffer is not an io.Closer; Close should be a no-op
	var buf bytes.Buffer
	w := NewSyncWriter(&buf)

	type closer interface{ Close() error }
	err := w.(closer).Close()
	assert.NoError(t, err)
}

func TestSyncWriterMutex(t *testing.T) {
	var buf bytes.Buffer
	mutex := new(sync.Mutex)
	w := NewSyncWriterMutex[*bytes.Buffer](&buf, mutex)

	n, err := w.Write([]byte("shared"))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "shared", buf.String())
}

func TestSyncWriterConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter[*bytes.Buffer](&buf)

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			_, _ = w.Write([]byte("x"))
		})
	}
	wg.Wait()
	assert.Equal(t, 50, buf.Len())
}

func TestSyncWriterConcurrentLockedAndWrite(t *testing.T) {
	var buf bytes.Buffer
	w := NewSyncWriter[*bytes.Buffer](&buf)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("w"))
		}()
		go func() {
			defer wg.Done()
			_ = w.Locked(func(b *bytes.Buffer) error {
				_, err := b.WriteString("l")
				return err
			})
		}()
	}
	wg.Wait()
	assert.Equal(t, 20, buf.Len())
}
