package util

import (
	"io"
	"sync"
)

// ReadFunc is the callback in NewFilterReader & NewFilterReadCloser for io.Reader.Read.
type ReadFunc func(bytes []byte) (n int, err error)

type filterReader struct {
	read ReadFunc
}

func (f *filterReader) Read(bytes []byte) (n int, err error) {
	if f.read == nil {
		err = io.EOF
		return
	}
	return f.read(bytes)
}

// NewFilterReader creates a new io.Reader redirects all read calls to ReadFunc
func NewFilterReader(read ReadFunc) io.Reader {
	return &filterReader{read}
}

// CloseFunc is the callback in NewFilterReadCloser for io.Closer.Close. It is guaranteed to be called once.
type CloseFunc func() (err error)

type filterReadCloser struct {
	filterReader
	close CloseFunc
}

func (c filterReadCloser) Close() error {
	defer func() {
		c.close = nil
		c.read = nil
	}()
	if c.close == nil {
		return nil
	}
	return c.close()
}

// NewFilterReadCloser creates a new io.ReadCloser redirects all calls to ReadFunc and CloseFunc
func NewFilterReadCloser(read ReadFunc, close CloseFunc) io.ReadCloser {
	return &filterReadCloser{*NewFilterReader(read).(*filterReader), close}
}

// NewSyncReader creates a new reader that is safe for concurrent use
func NewSyncReader[R io.Reader](reader R) SyncReader[R] {
	mutex := new(sync.Mutex)
	return NewSyncReaderMutex(reader, mutex, mutex)
}

// NewSyncReaderMutex creates a new reader that is safe for concurrent use and synced with the specified sync.Mutex
func NewSyncReaderMutex[R io.Reader](reader R, mutex, closeMutex *sync.Mutex) SyncReader[R] {
	return &syncReader[R]{reader: reader, mutex: mutex, closeMutex: closeMutex}
}

// SyncReader implements an io.ReadCloser that is safe for concurrent use
type SyncReader[R io.Reader] interface {
	io.ReadCloser
	// Locked provides sync.Mutex locked access to the underlying reader
	Locked(func(R) error) error
}

type syncReader[R io.Reader] struct {
	reader            io.Reader
	mutex, closeMutex *sync.Mutex
}

func (w *syncReader[R]) Locked(fn func(reader R) error) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return fn(w.reader.(R))
}

func (w *syncReader[R]) Read(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.reader.Read(p)
}

func (w *syncReader[R]) Close() (err error) {
	w.closeMutex.Lock()
	defer w.closeMutex.Unlock()
	if f, ok := w.reader.(io.Closer); ok {
		err = f.Close()
	}
	return
}
