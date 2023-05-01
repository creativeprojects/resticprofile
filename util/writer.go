package util

import (
	"io"
	"sync"
)

// Flusher allows a Writer to declare it may buffer content that can be flushed
type Flusher interface {
	// Flush writes any pending bytes to output
	Flush() error
}

// FlushWriter attempts to flush a writer if it implements Flusher
func FlushWriter(writer io.Writer) (flushable bool, err error) {
	var f Flusher
	if f, flushable = writer.(Flusher); flushable {
		err = f.Flush()
	}
	return
}

// NewSyncWriter creates a new writer that is safe for concurrent use
func NewSyncWriter[W io.Writer](writer W) SyncWriter[W] {
	return NewSyncWriterMutex(writer, new(sync.Mutex))
}

// NewSyncWriterMutex creates a new writer that is safe for concurrent use and synced with the specified sync.Mutex
func NewSyncWriterMutex[W io.Writer](writer W, mutex *sync.Mutex) SyncWriter[W] {
	return &syncWriter[W]{writer: writer, mutex: mutex}
}

// SyncWriter implements an io.Writer that is safe for concurrent use
type SyncWriter[W io.Writer] interface {
	io.Writer
	// Locked provides sync.Mutex locked access to the underlying writer
	Locked(func(W) error) error
}

type syncWriter[W io.Writer] struct {
	writer io.Writer
	mutex  *sync.Mutex
}

func (w *syncWriter[W]) Locked(fn func(writer W) error) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return fn(w.writer.(W))
}

func (w *syncWriter[W]) Write(p []byte) (n int, err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.writer.Write(p)
}

func (w *syncWriter[W]) Flush() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if f, ok := w.writer.(Flusher); ok {
		err = f.Flush()
	}
	return
}

func (w *syncWriter[W]) Close() (err error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if f, ok := w.writer.(io.Closer); ok {
		err = f.Close()
	}
	return
}
