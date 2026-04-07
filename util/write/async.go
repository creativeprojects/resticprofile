package write

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	asyncWriterDataChanSize  = 64
	asyncWriterFlushChanSize = 16
)

var ErrAlreadyClosed = errors.New("writer already closed")

type Async struct {
	handler       io.Writer
	interval      time.Duration
	data          chan []byte
	flusher       chan chan error
	done          chan struct{}
	systemGroup   sync.WaitGroup
	closeOnce     sync.Once
	closed        atomic.Bool
	flusherClosed atomic.Bool
}

// NewAsync creates a writer that accumulates Write requests and writes them at a fixed rate (every 250 ms by default)
func NewAsync(handler io.Writer, options ...AsyncOption) *Async {
	w := &Async{
		handler:  handler,
		interval: 250 * time.Millisecond,
		data:     make(chan []byte, asyncWriterDataChanSize),
		flusher:  make(chan chan error, asyncWriterFlushChanSize),
		done:     make(chan struct{}), // channel closed after the first call to Close()
	}
	for _, option := range options {
		option(w)
	}
	w.systemGroup.Go(func() {
		w.intervalFlush()
	})
	w.systemGroup.Go(func() {
		w.recvFlush()
	})
	return w
}

func (w *Async) intervalFlush() {
	ticker := time.NewTicker(w.interval)
	for {
		select {
		case <-ticker.C:
			w.flusher <- nil
		case <-w.done:
			ticker.Stop()
			return
		}
	}
}

func (w *Async) recvFlush() {
	for done := range w.flusher {
		err := w.flush()
		if done != nil { // some calls don't need to wait for the answer
			done <- err
		}
	}
}

// Close the writer. Any more call to Write will be ignored.
func (w *Async) Close() error {
	var err error
	w.closeOnce.Do(func() {
		w.closed.Store(true)
		close(w.done)
		err = w.Flush()
		w.flusherClosed.Store(true)
		close(w.flusher)
		w.systemGroup.Wait()
	})
	return err
}

func (w *Async) Flush() error {
	if w.flusherClosed.Load() {
		return fmt.Errorf("cannot write: %w", ErrAlreadyClosed)
	}
	done := make(chan error)
	w.flusher <- done
	// wait until the flusher is done
	err := <-done
	close(done)
	return err
}

func (w *Async) flush() error {
	for {
		// keep reading from the channel until it's empty
		select {
		case data := <-w.data:
			_, err := w.handler.Write(data)
			if err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// Write asynchronously to the handler
func (w *Async) Write(data []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, fmt.Errorf("cannot write: %w", ErrAlreadyClosed)
	}

	buffer := make([]byte, len(data))
	n = copy(buffer, data)
	w.data <- buffer
	return n, nil
}
