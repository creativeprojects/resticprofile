package util

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readCloser wraps a reader and tracks whether Close was called.
type trackingReadCloser struct {
	io.Reader
	closed bool
	err    error
}

func (t *trackingReadCloser) Close() error {
	t.closed = true
	return t.err
}

// filterReader & filterReadCloser tests

func TestFilterReaderRead(t *testing.T) {
	inner := strings.NewReader("hello")
	r := NewFilterReader(inner.Read)

	buf := make([]byte, 5)
	n, err := r.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", string(buf))
}

func TestFilterReaderReadEOF(t *testing.T) {
	inner := strings.NewReader("")
	r := NewFilterReader(inner.Read)

	buf := make([]byte, 4)
	_, err := r.Read(buf)
	assert.Equal(t, io.EOF, err)
}

func TestFilterReaderNilReadFunc(t *testing.T) {
	r := NewFilterReader(nil)

	buf := make([]byte, 4)
	n, err := r.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestFilterReaderReadAll(t *testing.T) {
	inner := strings.NewReader("read all")
	r := NewFilterReader(inner.Read)

	data, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, "read all", string(data))
}

func TestFilterReaderPropagatesError(t *testing.T) {
	sentinel := errors.New("read error")
	r := NewFilterReader(func([]byte) (int, error) { return 0, sentinel })

	buf := make([]byte, 4)
	_, err := r.Read(buf)
	assert.Equal(t, sentinel, err)
}

func TestFilterReadCloserRead(t *testing.T) {
	inner := strings.NewReader("world")
	rc := NewFilterReadCloser(inner.Read, func() error { return nil })

	buf := make([]byte, 5)
	n, err := rc.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "world", string(buf))
}

func TestFilterReadCloserClose(t *testing.T) {
	closed := false
	inner := strings.NewReader("")
	rc := NewFilterReadCloser(inner.Read, func() error {
		closed = true
		return nil
	})

	err := rc.Close()
	assert.NoError(t, err)
	assert.True(t, closed)
}

func TestFilterReadCloserCloseCalledTwice(t *testing.T) {
	calls := 0
	inner := strings.NewReader("")
	rc := NewFilterReadCloser(inner.Read, func() error {
		calls++
		return nil
	})

	_ = rc.Close()
	_ = rc.Close()
	assert.Equal(t, 1, calls)
}

func TestFilterReadCloserCloseError(t *testing.T) {
	sentinel := errors.New("close error")
	inner := strings.NewReader("")
	rc := NewFilterReadCloser(inner.Read, func() error { return sentinel })

	err := rc.Close()
	assert.Equal(t, sentinel, err)
}

func TestFilterReadCloserNilReadFunc(t *testing.T) {
	rc := NewFilterReadCloser(nil, func() error { return nil })

	buf := make([]byte, 4)
	n, err := rc.Read(buf)
	assert.Equal(t, io.EOF, err)
	assert.Equal(t, 0, n)
}

func TestFilterReadCloserNilCloseFunc(t *testing.T) {
	inner := strings.NewReader("")
	rc := NewFilterReadCloser(inner.Read, nil)

	err := rc.Close()
	assert.NoError(t, err)
}

func TestFilterReadCloserReadAfterClose(t *testing.T) {
	inner := strings.NewReader("data")
	rc := NewFilterReadCloser(inner.Read, func() error { return nil })

	require.NoError(t, rc.Close())

	buf := make([]byte, 4)
	n, err := rc.Read(buf)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

// syncReader tests

func TestNewSyncReaderRead(t *testing.T) {
	inner := strings.NewReader("hello world")
	sr := NewSyncReader[io.Reader](inner)

	buf := make([]byte, 11)
	n, err := sr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "hello world", string(buf))
}

func TestNewSyncReaderReadEOF(t *testing.T) {
	inner := strings.NewReader("")
	sr := NewSyncReader[io.Reader](inner)

	buf := make([]byte, 8)
	_, err := sr.Read(buf)
	assert.Equal(t, io.EOF, err)
}

func TestNewSyncReaderCloseWithCloser(t *testing.T) {
	rc := &trackingReadCloser{Reader: strings.NewReader("data")}
	sr := NewSyncReader(rc)

	err := sr.Close()
	assert.NoError(t, err)
	assert.True(t, rc.closed)
}

func TestNewSyncReaderCloseNonCloser(t *testing.T) {
	// strings.Reader does not implement io.Closer; Close should be a no-op
	inner := strings.NewReader("data")
	sr := NewSyncReader[io.Reader](inner)

	err := sr.Close()
	assert.NoError(t, err)
}

func TestNewSyncReaderCloseError(t *testing.T) {
	closeErr := errors.New("close failed")
	rc := &trackingReadCloser{Reader: strings.NewReader("data"), err: closeErr}
	sr := NewSyncReader(rc)

	err := sr.Close()
	assert.Equal(t, closeErr, err)
}

func TestNewSyncReaderLocked(t *testing.T) {
	inner := strings.NewReader("locked")
	sr := NewSyncReader(inner)

	var seen string
	err := sr.Locked(func(r *strings.Reader) error {
		buf := make([]byte, 6)
		n, err := r.Read(buf)
		seen = string(buf[:n])
		return err
	})
	assert.NoError(t, err)
	assert.Equal(t, "locked", seen)
}

func TestNewSyncReaderLockedPropagatesError(t *testing.T) {
	inner := strings.NewReader("data")
	sr := NewSyncReader(inner)

	sentinel := errors.New("locked error")
	err := sr.Locked(func(_ *strings.Reader) error {
		return sentinel
	})
	assert.Equal(t, sentinel, err)
}

func TestNewSyncReaderMutex(t *testing.T) {
	inner := strings.NewReader("shared mutex")
	mutex := new(sync.Mutex)
	sr := NewSyncReaderMutex[io.Reader](inner, mutex, mutex)

	buf := make([]byte, 12)
	n, err := sr.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 12, n)
	assert.Equal(t, "shared mutex", string(buf))
}

func TestNewSyncReaderMutexSeparateCloseMutex(t *testing.T) {
	rc := &trackingReadCloser{Reader: strings.NewReader("separate")}
	readMutex := new(sync.Mutex)
	closeMutex := new(sync.Mutex)
	sr := NewSyncReaderMutex(rc, readMutex, closeMutex)

	err := sr.Close()
	assert.NoError(t, err)
	assert.True(t, rc.closed)
}

func TestNewSyncReaderConcurrentReads(t *testing.T) {
	// Writes a large enough buffer so concurrent goroutines don't exhaust it
	// immediately; we just want to confirm no races / panics occur.
	data := strings.Repeat("x", 1024)
	inner := strings.NewReader(data)
	sr := NewSyncReader[io.Reader](inner)

	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			buf := make([]byte, 32)
			_, _ = sr.Read(buf)
		})
	}
	wg.Wait()
}

func TestNewSyncReaderConcurrentReadAndClose(t *testing.T) {
	rc := &trackingReadCloser{Reader: strings.NewReader(strings.Repeat("y", 512))}
	sr := NewSyncReader(rc)

	var wg sync.WaitGroup
	for range 5 {
		wg.Go(func() {
			buf := make([]byte, 16)
			_, _ = sr.Read(buf)
		})
	}
	wg.Go(func() {
		_ = sr.Close()
	})
	wg.Wait()
}

func TestNewSyncReaderReadAll(t *testing.T) {
	inner := strings.NewReader("read all content")
	sr := NewSyncReader[io.Reader](inner)

	content, err := io.ReadAll(sr)
	require.NoError(t, err)
	assert.Equal(t, "read all content", string(content))
}
